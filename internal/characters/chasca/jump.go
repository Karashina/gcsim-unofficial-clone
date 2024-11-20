package chasca

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
)

var skillJumpFrames []int

func init() {
	skillJumpFrames = frames.InitAbilSlice(54) // skill
	skillJumpFrames[action.ActionAttack] = 4
	skillJumpFrames[action.ActionBurst] = 50
	skillJumpFrames[action.ActionDash] = 49
	skillJumpFrames[action.ActionJump] = 50
	skillJumpFrames[action.ActionWalk] = 47
	skillJumpFrames[action.ActionSwap] = 48
}

func (c *char) Jump(p map[string]int) (action.Info, error) {
	if c.nightsoulState.HasBlessing() {
		c.reduceNightsoulPoints(21)

		return action.Info{
			Frames:          frames.NewAbilFunc(skillJumpFrames),
			AnimationLength: skillJumpFrames[action.InvalidAction],
			CanQueueAfter:   skillJumpFrames[action.ActionWalk],
			State:           action.JumpState,
		}, nil
	}
	return c.Character.Jump(p)
}
