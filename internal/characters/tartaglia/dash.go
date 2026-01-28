package tartaglia

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
)

var dashFrames []int

func init() {
	dashFrames = frames.InitAbilSlice(19)
	dashFrames[action.ActionSkill] = 3
}

func (c *char) Dash(p map[string]int) (action.Info, error) {
	c.Character.Dash(p)
	return action.Info{
		Frames:          func(next action.Action) int { return dashFrames[next] },
		AnimationLength: dashFrames[action.InvalidAction],
		CanQueueAfter:   dashFrames[action.ActionSkill], // fastest cancel
		State:           action.DashState,
	}, nil
}
