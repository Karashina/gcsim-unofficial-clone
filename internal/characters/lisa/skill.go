package lisa

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var (
	skillPressFrames []int
	skillHoldFrames  []int
)

const (
	skillPressHitmark = 22
	skillHoldHitmark  = 117
	particleICDKey    = "lisa-particle-icd"
)

func init() {
	// skill (press) -> x
	skillPressFrames = frames.InitAbilSlice(40)
	skillPressFrames[action.ActionAttack] = 37
	skillPressFrames[action.ActionCharge] = 38
	skillPressFrames[action.ActionDash] = 35
	skillPressFrames[action.ActionJump] = 20
	skillPressFrames[action.ActionSwap] = 20

	// skill (hold) -> x
	skillHoldFrames = frames.InitAbilSlice(143)
	skillHoldFrames[action.ActionCharge] = 138
	skillHoldFrames[action.ActionBurst] = 138
	skillHoldFrames[action.ActionDash] = 116
	skillHoldFrames[action.ActionJump] = 117
	skillHoldFrames[action.ActionSwap] = 117
}

// p = 0: 短押し、p = 1: 長押し
func (c *char) Skill(p map[string]int) (action.Info, error) {
	hold := p["hold"]
	if hold == 1 {
		return c.skillHold(), nil
	}
	return c.skillPress(), nil
}

func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, 0.2*60, true)
	c.Core.QueueParticle(c.Base.Key.String(), 5, attributes.Electro, c.ParticleDelay)
}

// TODO: スタックの持続時間は？
func (c *char) skillPress() action.Info {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Violet Arc",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagLisaElectro,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       skillPress[c.TalentLvlSkill()],
	}

	cb := func(a combat.AttackCB) {
		// オフフィールドではスタックが蓄積されない
		if c.Core.Player.Active() != c.Index {
			return
		}
		t, ok := a.Target.(*enemy.Enemy)
		if !ok {
			return
		}
		count := t.GetTag(conductiveTag)
		if count < 3 {
			t.SetTag(conductiveTag, count+1)
		}
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHit(c.Core.Combat.Player(), c.Core.Combat.PrimaryTarget(), nil, 1),
		0,
		skillPressHitmark,
		cb,
	)

	c.SetCDWithDelay(action.ActionSkill, 60, 17)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillPressFrames),
		AnimationLength: skillPressFrames[action.InvalidAction],
		CanQueueAfter:   skillPressFrames[action.ActionSwap], // 最速キャンセル
		State:           action.SkillState,
	}
}

// 長時間の詠唱後、天から雷を呼び降ろし、周囲の敵に大規模な雷元素ダメージを与える。
// 導電スタックの数に応じて大きな追加ダメージを与え、導電状態を解除する。
func (c *char) skillHold() action.Info {
	// 倍率なし（ターゲット依存のため）
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Violet Arc (Hold)",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 50,
	}

	// 2凸: 防御力上昇。いずれにせよ中断なし
	if c.Base.Cons >= 2 {
		// このアビリティの持続フレーム中、防御力を上昇させる
		m := make([]float64, attributes.EndStatType)
		m[attributes.DEFP] = 0.25
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("lisa-c2", 126),
			AffectedStat: attributes.DEFP,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}

	count := 0
	var c1cb func(a combat.AttackCB)
	if c.Base.Cons > 0 {
		c1cb = func(a combat.AttackCB) {
			if a.Target.Type() != targets.TargettableEnemy {
				return
			}
			if count == 5 {
				return
			}
			count++
			c.QueueCharTask(func() {
				c.AddEnergy("lisa-c1", 2)
			}, 0.4*60)
		}
	}

	// [8:31 PM] ArchedNosi | Lisa Unleashed: 長押しで常に5個のオーブを生成
	enemies := c.Core.Combat.EnemiesWithinArea(combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 10), nil)
	for _, e := range enemies {
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(e, nil, 0.2), 0, skillHoldHitmark, c1cb, c.particleCB)
	}

	c.SetCDWithDelay(action.ActionSkill, 960, 114)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillHoldFrames),
		AnimationLength: skillHoldFrames[action.InvalidAction],
		CanQueueAfter:   skillHoldFrames[action.ActionDash], // 最速キャンセル
		State:           action.SkillState,
	}
}

func (c *char) skillHoldMult() {
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		t, ok := args[0].(*enemy.Enemy)
		if !ok {
			return false
		}
		if atk.Info.Abil != "Violet Arc (Hold)" {
			return false
		}
		stacks := t.GetTag(conductiveTag)

		atk.Info.Mult = skillHold[stacks][c.TalentLvlSkill()]

		// スタックを消費
		t.SetTag(conductiveTag, 0)

		return false
	}, "lisa-skill-hold-mul")
}
