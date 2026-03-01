package varka

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var burstFrames []int

const burstHitmark1 = 114
const burstHitmark2 = 134

func init() {
	burstFrames = frames.InitAbilSlice(143)
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	lvl := c.TalentLvlBurst()

	// 1st Hit: converts to Other element if eligible, otherwise Anemo
	ele1 := attributes.Anemo
	if c.hasOtherEle {
		ele1 = c.otherElement
	}

	ai1 := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Northwind Avatar: 1st Hit",
		AttackTag:          attacks.AttackTagElementalBurst,
		ICDTag:             attacks.ICDTagNone,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeBlunt,
		PoiseDMG:           200.0,
		Element:            ele1,
		Durability:         25,
		Mult:               burst1[lvl],
		HitlagFactor:       0.01,
		HitlagHaltFrames:   0.12 * 60,
		CanBeDefenseHalted: true,
	}
	c.Core.QueueAttack(
		ai1,
		combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, 6, 8),
		burstHitmark1, burstHitmark1,
	)

	// 2nd Hit: always Anemo
	ai2 := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Northwind Avatar: 2nd Hit",
		AttackTag:          attacks.AttackTagElementalBurst,
		ICDTag:             attacks.ICDTagNone,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeBlunt,
		PoiseDMG:           250.0,
		Element:            attributes.Anemo,
		Durability:         25,
		Mult:               burst2[lvl],
		HitlagFactor:       0.01,
		HitlagHaltFrames:   0.15 * 60,
		CanBeDefenseHalted: true,
	}
	c.Core.QueueAttack(
		ai2,
		combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, 6, 8),
		burstHitmark2, burstHitmark2,
	)

	// CD 15s, Energy cost 60
	c.SetCD(action.ActionBurst, 15*60)
	c.ConsumeEnergy(3) // frame delay for energy drain

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstHitmark1,
		State:           action.BurstState,
	}, nil
}
