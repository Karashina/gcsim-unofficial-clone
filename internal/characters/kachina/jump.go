package kachina

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
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
	if c.StatusIsActive(skillRideKey) {
		return c.AttackRide(32)
	}

	ai, err := c.Character.Jump(p)

	f := ai.AnimationLength
	ai.Frames = func(action.Action) int { return f }
	ai.AnimationLength = f
	ai.CanQueueAfter = f

	return ai, err
}
