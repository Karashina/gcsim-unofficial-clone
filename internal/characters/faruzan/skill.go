package faruzan

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var skillFrames []int

const (
	VortexAbilName = "Pressurized Collapse"
	skillHitmark   = 14
	skillKey       = "faruzan-e"
	particleICDKey = "faruzan-particle-icd"
	vortexHitmark  = 33
)

func init() {
	skillFrames = frames.InitAbilSlice(35)
	skillFrames[action.ActionSkill] = 34 // E -> E
	skillFrames[action.ActionBurst] = 34 // E -> Q
	skillFrames[action.ActionDash] = 28  // E -> N1
	skillFrames[action.ActionJump] = 27  // E -> J
	skillFrames[action.ActionWalk] = 34  // E -> J
	skillFrames[action.ActionSwap] = 33  // E -> Swap
}

// ファルザンが多面体を展開し、付近の敵に範囲風元素ダメージを与える。
// 同時に「風導」状態に入る。「風導」状態では次のフルチャージ狙い撃ちが
// この状態を消費し、風元素ダメージを与えるハリケーンアローとなる。
// このダメージは重撃ダメージとみなされる。
//
// 結圧崩壊
// ハリケーンアローは着弾点に結圧崩壊効果を生成し、
// 命中した敵またはキャラクターに結圧崩壊効果を適用する。
// この効果は短時間後に解除され、範囲風元素ダメージを与え、
// 付近のオブジェクトと敵を引き寄せる渦巻きを生成する。
// 渦巻きのダメージは元素スキルダメージとみなされる。
func (c *char) Skill(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Wind Realm of Nasamjnin (E)",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       skill[c.TalentLvlSkill()],
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 3}, 3),
		skillHitmark,
		skillHitmark,
	)

	// 1凸: 風域の効果中、フルチャージ狙い撃ちで
	// ハリケーンアローを最大2回発射可能。
	c.hurricaneCount = 1
	if c.Base.Cons >= 1 {
		c.hurricaneCount = 2
	}

	c.Core.Tasks.Add(func() {
		c.AddStatus(skillKey, 1080, true)
		c.SetCD(action.ActionSkill, 360)
	}, 12)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionJump], // 最速キャンセル
		State:           action.SkillState,
	}, nil
}

func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, 5.5*60, true)
	c.Core.QueueParticle(c.Base.Key.String(), 2, attributes.Anemo, c.ParticleDelay)
}

func (c *char) pressurizedCollapse(pos geometry.Point) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       VortexAbilName,
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       vortexDmg[c.TalentLvlSkill()],
	}
	snap := c.Snapshot(&ai)

	// 固有天賦1:
	// 結圧崩壊の渦巻きに当たった敵に秘羽の虎風の効果を付与できる。
	// 彼女は結圧崩壊の渦巻きに命中した敵に秘羽の虎風の結圧の悪風を付与できる。
	var shredCb combat.AttackCBFunc
	if c.Base.Ascension >= 1 {
		shredCb = applyBurstShredCb
	}

	c.Core.Tasks.Add(func() {
		c.Core.QueueAttackWithSnap(
			ai,
			snap,
			combat.NewCircleHitOnTarget(pos, nil, 6),
			0,
			c.makeC4Callback(),
			shredCb,
			c.particleCB,
		)
	}, vortexHitmark)
}
