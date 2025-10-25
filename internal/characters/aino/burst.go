package aino

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
)

var burstFrames []int

const (
	burstDuration       = 14 * 60 // 14 seconds
	burstTickInterval   = 2 * 60  // 2 seconds (default)
	burstTickIntervalA1 = 1 * 60  // 1 second (when Moonsign is Ascendant)
	burstKey            = "aino-burst"
)

func init() {
	burstFrames = frames.InitAbilSlice(95)
	burstFrames[action.ActionAttack] = 85
	burstFrames[action.ActionSkill] = 85
	burstFrames[action.ActionDash] = 75
	burstFrames[action.ActionSwap] = 82
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	c.burstSrc = c.Core.F

	// Determine interval based on A1 passive
	interval := burstTickInterval
	radius := 3.0
	if c.MoonsignAscendant {
		interval = burstTickIntervalA1
		radius = 5.0
	}

	// Schedule water ball attacks
	numTicks := burstDuration / interval
	for i := 0; i < numTicks; i++ {
		c.Core.Tasks.Add(c.burstTick(c.burstSrc, radius), interval*i)
	}

	c.AddStatus(burstKey, burstDuration, false)

	c.SetCD(action.ActionBurst, 13.5*60) // 13.5s cooldown
	c.ConsumeEnergy(7)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash],
		State:           action.BurstState,
	}, nil
}

func (c *char) burstTick(src int, radius float64) func() {
	return func() {
		if src != c.burstSrc {
			return
		}
		if !c.StatusIsActive(burstKey) {
			return
		}

		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Precision Hydronic Cooler (Water Ball)",
			AttackTag:  attacks.AttackTagElementalBurst,
			ICDTag:     attacks.ICDTagElementalBurst,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Hydro,
			Durability: 25,
			Mult:       burst[c.TalentLvlBurst()],
			FlatDmg:    c.a4FlatDmgBuff,
		}

		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, radius),
			0,
			0,
		)
	}
}
