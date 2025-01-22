package lanyan

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
)

const (
	burstCDDelay       = 1
	EnergyConsumeDelay = 4
)

var (
	burstHitmarks = []int{31, 40, 47}
	burstFrames   []int
)

func init() {
	burstFrames = frames.InitAbilSlice(80) // Q > D
	burstFrames[action.ActionAttack] = 75  // Q > NA
	burstFrames[action.ActionSkill] = 78   // Q > E
	burstFrames[action.ActionJump] = 72    // Q > J
}

func (c *char) Burst(p map[string]int) (action.Info, error) {

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Lustrous Moonrise",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 50,
		Mult:       burst[c.TalentLvlBurst()],
	}
	// A4 Flat DMG
	if c.Base.Ascension >= 4 {
		ai.FlatDmg = 7.74 * c.Stat(attributes.EM)
	}
	for i := 0; i < 3; i++ {
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 6), 0, burstHitmarks[i])
	}

	c.SetCDWithDelay(action.ActionBurst, 15*60, burstCDDelay)
	c.ConsumeEnergy(EnergyConsumeDelay)
	c.c4()

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // earliest cancel
		State:           action.BurstState,
	}, nil
}
