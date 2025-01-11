package pyro

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
)

var (
	burstHitmark           = 37
	consumeEnergyFrame     = 4
	generateNightsoulFrame = 44

	burstFrames [][]int
)

const burstKey = "traveller-burst"

func init() {
	burstFrames = make([][]int, 2)

	// Male
	burstFrames[0] = burstFrames[1] // Q -> E/D/Walk

	// Female
	burstFrames[1] = frames.InitAbilSlice(78) // Q -> Walk
}

func (c *Traveler) Burst(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Plains Scorcher",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupTravelerBurst,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Pyro,
		Durability: 25,
		Mult:       burst[c.TalentLvlBurst()],
	}

	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player().Pos(), nil, 5), burstHitmark, burstHitmark)
	c.SetCD(action.ActionBurst, 18*60)
	c.c4()
	c.AddStatus(burstKey, 239+generateNightsoulFrame, false)
	c.QueueCharTask(c.gainNightsoul(), generateNightsoulFrame)
	c.ConsumeEnergy(consumeEnergyFrame)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames[c.gender]),
		AnimationLength: burstFrames[c.gender][action.InvalidAction],
		CanQueueAfter:   burstFrames[c.gender][action.ActionSwap], // earliest cancel
		State:           action.BurstState,
	}, nil
}

func (c *Traveler) gainNightsoul() func() {
	return func() {
		if !c.StatusIsActive(burstKey) {
			return
		}
		c.nightsoulState.GeneratePoints(7)
		c.QueueCharTask(c.gainNightsoul(), 60)
	}
}
