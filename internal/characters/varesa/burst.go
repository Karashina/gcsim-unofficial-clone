package varesa

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
)

var (
	burstFrames    []int
	volcanicFrames []int
)

const (
	burstHitmark     = 90
	burstEnergyFrame = 10

	kablamHitmark = 41
	kablamCost    = 30
	kablamAbil    = "Volcano Kablam"
	c4Key         = "varesa-c4"
)

func init() {
	burstFrames = frames.InitAbilSlice(103)

	volcanicFrames = frames.InitAbilSlice(46)
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	if c.Base.Cons >= 4 && !c.StatusIsActive(apexDriveKey) && !c.StatusIsActive(fieryPassionKey) {
		c.AddStatus(c4Key, 15*60, true)
	}
	if c.StatusIsActive(apexDriveKey) {
		c.DeleteStatus(apexDriveKey)
		return c.volcanicKablam(), nil
	}
	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Flying Kick",
		AttackTag:      attacks.AttackTagElementalBurst,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagNone,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Electro,
		Durability:     25,
		Mult:           burst[c.TalentLvlBurst()],
	}

	if c.nightsoulState.HasBlessing() {
		ai.Abil = "Flying Kick (Fiery Passion)"
		ai.Mult = burst[c.TalentLvlBurst()]
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHit(c.Core.Combat.Player(), c.Core.Combat.PrimaryTarget(), nil, 6),
		burstHitmark,
		burstHitmark,
	)

	c.ConsumeEnergy(burstEnergyFrame)
	c.SetCD(action.ActionBurst, 18*60)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap],
		State:           action.BurstState,
	}, nil
}

func (c *char) volcanicKablam() action.Info {
	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           kablamAbil,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		AttackTag:      attacks.AttackTagPlunge,
		ICDTag:         attacks.ICDTagVaresaCombatCycle,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Electro,
		Durability:     25,
		Mult:           volcano[c.TalentLvlBurst()],
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHit(c.Core.Combat.Player(), c.Core.Combat.PrimaryTarget(), nil, 6),
		kablamHitmark,
		kablamHitmark,
	)

	c.AddEnergy("varesa-kablam", -kablamCost)
	c.SetCD(action.ActionBurst, 1*60)

	return action.Info{
		Frames:          frames.NewAbilFunc(volcanicFrames),
		AnimationLength: volcanicFrames[action.InvalidAction],
		CanQueueAfter:   volcanicFrames[action.ActionSwap],
		State:           action.BurstState,
	}
}
