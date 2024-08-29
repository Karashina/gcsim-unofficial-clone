package mualani

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
)

var EJumpFrames []int

func init() {
	EJumpFrames = frames.InitAbilSlice(53)
	EJumpFrames[action.ActionAttack] = 53
	EJumpFrames[action.ActionCharge] = 53
	EJumpFrames[action.ActionSkill] = 53
	EJumpFrames[action.ActionBurst] = 53
	EJumpFrames[action.ActionDash] = 53
	EJumpFrames[action.ActionWalk] = 53
}

func (c *char) Jump(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(skillKey) {
		return c.WindfavoredJump(p)
	}

	ai, err := c.Character.Jump(p)

	f := ai.AnimationLength
	ai.Frames = func(action.Action) int { return f }
	ai.AnimationLength = f
	ai.CanQueueAfter = f

	return ai, err
}

func (c *char) WindfavoredJump(p map[string]int) (action.Info, error) {
	c.ConsumeNightsoul(2)
	return action.Info{
		Frames:          frames.NewAbilFunc(EJumpFrames),
		AnimationLength: EJumpFrames[action.ActionJump],
		CanQueueAfter:   EJumpFrames[action.ActionSkill], // earliest cancel
		State:           action.JumpState,
	}, nil
}
