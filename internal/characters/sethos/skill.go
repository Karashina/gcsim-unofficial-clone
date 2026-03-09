package sethos

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var skillFrames []int

const skillParticleICDKey = "sethos-particle-icd"

func init() {
	skillFrames = frames.InitAbilSlice(38) // E -> Charge
	skillFrames[action.ActionAttack] = 28
	skillFrames[action.ActionSkill] = 32
	skillFrames[action.ActionBurst] = 28
	skillFrames[action.ActionDash] = 27
	skillFrames[action.ActionJump] = 27
	skillFrames[action.ActionWalk] = 25
	skillFrames[action.ActionSwap] = 30
}

func (c *char) skillRefundHook() {
	refundCB := func(args ...interface{}) bool {
		// TODO: SethosのEスキルが敵でフィルタリングするか確認
		// a := args[0].(combat.Target)
		// if a.Type() != targets.TargettableEnemy {
		// 	falseを返す
		// }
		ae := args[1].(*combat.AttackEvent)
		if ae.Info.ActorIndex != c.Index {
			return false
		}
		if ae.Info.AttackTag != attacks.AttackTagElementalArt {
			return false
		}
		// 範囲攻撃で二重発動を防止
		if c.lastSkillFrame == ae.SourceFrame {
			return false
		}
		c.lastSkillFrame = ae.SourceFrame
		c.AddEnergy("sethos-skill", skillEnergyRegen[c.TalentLvlSkill()])
		c.c2AddStack(c2RegainingKey)

		return false
	}

	c.Core.Events.Subscribe(event.OnOverload, refundCB, "sethos-e-refund")
	c.Core.Events.Subscribe(event.OnElectroCharged, refundCB, "sethos-e-refund")
	c.Core.Events.Subscribe(event.OnSuperconduct, refundCB, "sethos-e-refund")
	c.Core.Events.Subscribe(event.OnSwirlElectro, refundCB, "sethos-e-refund")
	c.Core.Events.Subscribe(event.OnHyperbloom, refundCB, "sethos-e-refund")
	c.Core.Events.Subscribe(event.OnQuicken, refundCB, "sethos-e-refund")
	c.Core.Events.Subscribe(event.OnAggravate, refundCB, "sethos-e-refund")
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Ancient Rite: Thunderous Roar of Sand",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       skill[c.TalentLvlSkill()],
	}

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 4.5)
	c.Core.QueueAttack(ai, ap, 0, 13, c.particleCB)

	c.SetCDWithDelay(action.ActionSkill, 8*60, 10)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionSwap], // 最速キャンセル
		State:           action.SkillState,
	}, nil
}

func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(skillParticleICDKey) {
		return
	}
	c.AddStatus(skillParticleICDKey, 0.5*60, true)
	c.Core.QueueParticle(c.Base.Key.String(), 2, attributes.Electro, c.ParticleDelay)
}
