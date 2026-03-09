package xianyun

import (
	"slices"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var skillLeapFrames [][]int
var skillStateDur = []int{220, 238, 179}

const (
	skillPressHitmark = 3

	skillStateKey = "cloud-transmogrification"

	// TODO: スキルの判定範囲を調査。現在は落下攻撃の衝突判定と同じサイズと仮定
	skillRadius = 1.5

	particleCount  = 5
	particleICD    = 0.2 * 60
	particleICDKey = "xianyun-particle-icd"
)

func init() {
	skillLeapFrames = make([][]int, 3)
	skillLeapFrames[0] = frames.InitAbilSlice(244)

	skillLeapFrames[0][action.ActionAttack] = 221
	skillLeapFrames[0][action.ActionSkill] = 14
	skillLeapFrames[0][action.ActionBurst] = 40
	skillLeapFrames[0][action.ActionDash] = 39
	skillLeapFrames[0][action.ActionJump] = 41
	skillLeapFrames[0][action.ActionWalk] = 43
	skillLeapFrames[0][action.ActionSwap] = 36
	skillLeapFrames[0][action.ActionHighPlunge] = 14
	skillLeapFrames[0][action.ActionLowPlunge] = 14

	skillLeapFrames[1] = frames.InitAbilSlice(243)
	skillLeapFrames[1][action.ActionSkill] = 15
	skillLeapFrames[1][action.ActionBurst] = 60
	skillLeapFrames[1][action.ActionDash] = 60
	skillLeapFrames[1][action.ActionJump] = 60
	skillLeapFrames[1][action.ActionWalk] = 66
	skillLeapFrames[1][action.ActionSwap] = 59
	skillLeapFrames[1][action.ActionHighPlunge] = 15
	skillLeapFrames[1][action.ActionLowPlunge] = 15

	skillLeapFrames[2] = frames.InitAbilSlice(178)
	skillLeapFrames[2][action.ActionSkill] = 128
	skillLeapFrames[2][action.ActionBurst] = 126
	skillLeapFrames[2][action.ActionDash] = 130
	skillLeapFrames[2][action.ActionJump] = 129
	skillLeapFrames[2][action.ActionWalk] = 125
	skillLeapFrames[2][action.ActionSwap] = 126
	skillLeapFrames[2][action.ActionHighPlunge] = 18
	skillLeapFrames[2][action.ActionLowPlunge] = 18
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	// 最初の跳躍を確認
	if !c.StatusIsActive(skillStateKey) || c.skillCounter == 3 { // Didn't plunge after the previous triple skill
		c.skillCounter = 0
		if c.StatusIsActive(c6Key) {
			c.skillWasC6 = true
			c.SetTag(c6Key, c.Tag(c6Key)-1)
			if c.Tag(c6Key) <= 0 {
				c.DeleteStatus(c6Key)
			}
		} else {
			c.SetCD(action.ActionSkill, 12*60)
			c.skillWasC6 = false
		}
		c.skillEnemiesHit = nil
	}
	// 2凸: 白雲の暁を使用後、閑雲の攻撃力が15秒間20%上昇する。
	c.c2buff()

	// 各敵に対して最大1回のみヒットする
	// 閑雲が雲変化状態に入るたびに、天梯は最大3回使用でき、各敵に対して1回分の天梯ダメージのみ与えられる。
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Skyladder",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 0,
		Mult:       skillPress[c.TalentLvlSkill()],
	}

	aoe := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, skillRadius)
	targets := c.Core.Combat.EnemiesWithinArea(aoe, func(t combat.Enemy) bool {
		return !slices.Contains[[]targets.TargetKey](c.skillEnemiesHit, t.Key())
	})

	for _, t := range targets {
		c.Core.QueueAttack(
			ai,
			combat.NewSingleTargetHit(t.Key()),
			skillPressHitmark,
			skillPressHitmark,
		)
		c.skillEnemiesHit = append(c.skillEnemiesHit, t.Key())
	}

	c.skillSrc = c.Core.F
	c.QueueCharTask(c.cooldownReduce(c.Core.F), skillStateDur[c.skillCounter])
	c.AddStatus(skillStateKey, skillStateDur[c.skillCounter], true)

	defer func() { c.skillCounter++ }()

	return action.Info{
		Frames:          frames.NewAbilFunc(skillLeapFrames[c.skillCounter]),
		AnimationLength: skillLeapFrames[c.skillCounter][action.InvalidAction],
		CanQueueAfter:   skillLeapFrames[c.skillCounter][action.ActionHighPlunge], // 最速キャンセル
		State:           action.SkillState,
	}, nil
}

func (c *char) cooldownReduce(src int) func() {
	return func() {
		if c.skillSrc != src {
			return
		}
		// この状態中に鶴雲波を使用しなかった場合、白雲の暁の次のCDが3秒短縮される。
		c.ReduceActionCooldown(action.ActionSkill, 3*60)
	}
}

func (c *char) particleCB() func(combat.AttackCB) {
	// 6凸由来のスキルの場合、粒子は生成されない
	if c.skillWasC6 {
		return nil
	}
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if c.StatusIsActive(particleICDKey) {
			return
		}
		c.AddStatus(particleICDKey, particleICD, true)

		c.Core.QueueParticle(c.Base.Key.String(), particleCount, attributes.Anemo, c.ParticleDelay)
	}
}
