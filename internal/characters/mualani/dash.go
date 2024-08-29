package mualani

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
)

var (
	dashFramesE []int
)

func init() {
	dashFramesE = frames.InitAbilSlice(65)
	dashFramesE[action.ActionAttack] = 65
	dashFramesE[action.ActionCharge] = 65
	dashFramesE[action.ActionSkill] = 65
	dashFramesE[action.ActionDash] = 65
	dashFramesE[action.ActionJump] = 65
	dashFramesE[action.ActionWalk] = 65
}

func (c *char) Dash(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(skillKey) {
		return c.NightsoulDash(p)
	}

	c.ApplyDashCD()
	c.QueueDashStaminaConsumption(p)
	length := c.DashLength()
	return action.Info{
		Frames:          func(action.Action) int { return length },
		AnimationLength: length,
		CanQueueAfter:   length,
		State:           action.DashState,
	}, nil
}

func (c *char) NightsoulDash(p map[string]int) (action.Info, error) {
	ai := action.Info{
		Frames:          func(next action.Action) int { return dashFramesE[next] },
		AnimationLength: dashFramesE[action.InvalidAction],
		CanQueueAfter:   dashFramesE[action.ActionSkill],
		State:           action.DashState,
	}
	c.ConsumeNightsoul(10)

	return ai, nil
}
