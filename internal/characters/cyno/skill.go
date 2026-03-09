package cyno

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

const (
	skillBName     = "Mortuary Rite"
	particleICDKey = "cyno-particle-icd"
)

var (
	skillCD       = 450
	skillBCD      = 180
	skillCDDelay  = 17
	skillBCDDelay = 26
	skillHitmark  = 21
	skillBHitmark = 28
	skillFrames   []int
	skillBFrames  []int
)

func init() {
	skillFrames = frames.InitAbilSlice(43)
	skillFrames[action.ActionDash] = 31
	skillFrames[action.ActionJump] = 32
	skillFrames[action.ActionSwap] = 42

	// 元素爆発中のフレーム
	skillBFrames = frames.InitAbilSlice(34)
	skillBFrames[action.ActionDash] = 30
	skillBFrames[action.ActionJump] = 31
	skillBFrames[action.ActionSwap] = 33
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(burstKey) {
		return c.skillB()
	}

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Secret Rite: Chasmic Soulfarer",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeSlash,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       skill[c.TalentLvlSkill()],
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHit(
			c.Core.Combat.Player(),
			c.Core.Combat.PrimaryTarget(),
			nil,
			1,
		),
		skillHitmark,
		skillHitmark,
		c.makeParticleCB(false),
	)

	c.Core.Tasks.Add(c.triggerSkillCD, skillCDDelay)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash], // 最速キャンセル
		State:           action.SkillState,
	}, nil
}

func (c *char) skillB() (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex:       c.Index,
		Abil:             skillBName,
		AttackTag:        attacks.AttackTagElementalArt,
		ICDTag:           attacks.ICDTagNone,
		ICDGroup:         attacks.ICDGroupDefault,
		StrikeType:       attacks.StrikeTypeBlunt,
		PoiseDMG:         75,
		Element:          attributes.Electro,
		Durability:       25,
		Mult:             skillB[c.TalentLvlSkill()],
		HitlagFactor:     0.01,
		HitlagHaltFrames: 0.03 * 60,
	}

	ap := combat.NewCircleHitOnTarget(
		c.Core.Combat.Player(),
		geometry.Point{Y: 1.5},
		6,
	)
	particleCB := c.makeParticleCB(true)
	if !c.StatusIsActive(a1Key) { // 「末途見」バフをチェック
		c.Core.QueueAttack(ai, ap, skillBHitmark, skillBHitmark, particleCB)
	} else {
		// 元素スキルの追加ダメージを適用
		c.a1Buff()
		if c.Base.Cons >= 1 && c.StatusIsActive(c1Key) {
			c.c1()
		}
		c.c6Init()

		c.Core.QueueAttack(ai, ap, skillBHitmark, skillBHitmark, particleCB)
		// 元素スキルの追加ダメージを適用
		ai.Abil = "Duststalker Bolt"
		ai.Mult = 1.0
		ai.FlatDmg = c.a4Bolt()
		ai.AttackTag = attacks.AttackTagElementalArtHold
		ai.ICDTag = attacks.ICDTagElementalArt
		ai.ICDGroup = attacks.ICDGroupCynoBolt
		ai.StrikeType = attacks.StrikeTypeSlash
		ai.PoiseDMG = 25
		ai.HitlagFactor = 0
		ai.HitlagHaltFrames = 0

		// 3インスタンス
		for i := 0; i < 3; i++ {
			c.Core.QueueAttack(
				ai,
				combat.NewCircleHit(
					c.Core.Combat.Player(),
					c.Core.Combat.PrimaryTarget(),
					nil,
					0.3,
				),
				skillBHitmark,
				skillBHitmark,
				particleCB,
			)
		}
	}
	if c.burstExtension < 2 { // 元素爆発ど1回につき最大2回延長可能（基本10秒 + 每回4秒で最大18秒）
		c.ExtendStatus(burstKey, 240) // 4s*60
		c.burstExtension++
	}

	c.tryBurstPPSlide(skillBHitmark)

	c.Core.Tasks.Add(c.triggerSkillCD, skillBCDDelay)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillBFrames),
		AnimationLength: skillBFrames[action.InvalidAction],
		CanQueueAfter:   skillBFrames[action.ActionDash], // 最速キャンセル
		State:           action.SkillState,
	}, nil
}

func (c *char) triggerSkillCD() {
	c.ResetActionCooldown(action.ActionSkill)
	c.SetCD(action.ActionSkill, skillCD)
	c.SetCD(action.ActionLowPlunge, skillBCD)
}

func (c *char) makeParticleCB(burst bool) combat.AttackCBFunc {
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if c.StatusIsActive(particleICDKey) {
			return
		}
		c.AddStatus(particleICDKey, 0.5*60, true)

		var count float64
		if burst {
			count = 1
			if c.Core.Rand.Float64() < 0.33 {
				count = 2
			}
		} else {
			count = 3
		}
		c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Electro, c.ParticleDelay)
	}
}
