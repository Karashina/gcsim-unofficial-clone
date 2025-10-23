package nefer

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
)

var (
	burstFrames []int
)

const (
	burstHitmark = 100
)

func init() {
	burstFrames = frames.InitAbilSlice(115)
}

// Elemental Burst
func (c *char) Burst(p map[string]int) (action.Info, error) {
	/*
	   deals 2 AoE Dendro DMG.
	*/

	c.SetCD(action.ActionBurst, 15*60)
	c.ConsumeEnergy(5)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap],
		State:           action.BurstState,
	}, nil
}

// When unleashed, Nefer will consume all Veils of Falsehood to increase the DMG dealt by the current Elemental Burst.
// TO COPILOT: buff is "burstbonus", implement as normal AttackMod
