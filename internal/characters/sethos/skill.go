package sethos

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/targets"
	"github.com/genshinsim/gcsim/pkg/gadget"
)

var skillFrames []int

const (
	skillHitmark   = 14
	particleICDKey = "sethos-particle-icd"
)

func init() {
	skillFrames = frames.InitAbilSlice(31)
	skillFrames[action.ActionAttack] = 31
	skillFrames[action.ActionAim] = 31
	skillFrames[action.ActionBurst] = 31
	skillFrames[action.ActionDash] = 31
	skillFrames[action.ActionJump] = 31
	skillFrames[action.ActionSwap] = 31
}

func (c *char) Skill(p map[string]int) (action.Info, error) {

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Ancient Rite: The Thundering Sands(E)",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       skill[c.TalentLvlSkill()],
	}
	c.skillArea = combat.NewCircleHit(c.Core.Combat.Player(), c.Core.Combat.PrimaryTarget(), nil, 4.5)
	c.Core.QueueAttack(ai, c.skillArea, skillHitmark, skillHitmark, c.particleCB)

	c.SetCDWithDelay(action.ActionSkill, 8*60, 0)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionAim], // earliest cancel
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
	c.AddStatus(particleICDKey, 0.3*60, true)

	count := 2.0
	c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Electro, c.ParticleDelay)
}

func (c *char) Energygen() {
	energycb := func(args ...interface{}) bool {
		if c.Core.Player.Active() != c.Index {
			return false
		}
		if args[1].(*combat.AttackEvent).Info.AttackTag != attacks.AttackTagElementalArt {
			return false
		}
		c.AddEnergy("Sethos-skill-energy", 12)
		if c.Base.Cons >= 2 {
			c.AddStatus(c2energykey, 600, true)
		}
		return false
	}

	energycbNoGadget := func(args ...interface{}) bool {
		if _, ok := args[0].(*gadget.Gadget); ok {
			return false
		}

		return energycb(args...)
	}

	c.Core.Events.Subscribe(event.OnOverload, energycbNoGadget, "sethos-a4")
	c.Core.Events.Subscribe(event.OnElectroCharged, energycbNoGadget, "sethos-a4")
	c.Core.Events.Subscribe(event.OnSuperconduct, energycbNoGadget, "sethos-a4")
	c.Core.Events.Subscribe(event.OnSwirlElectro, energycbNoGadget, "sethos-a4")
	c.Core.Events.Subscribe(event.OnHyperbloom, energycb, "sethos-a4")
	c.Core.Events.Subscribe(event.OnQuicken, energycbNoGadget, "sethos-a4")
	c.Core.Events.Subscribe(event.OnAggravate, energycbNoGadget, "sethos-a4")
}
