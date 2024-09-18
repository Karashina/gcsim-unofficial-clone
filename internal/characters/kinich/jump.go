package kinich

import (
	"errors"

	"github.com/genshinsim/gcsim/pkg/core/action"
)

var EJumpFrames []int

func (c *char) Jump(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(skillLinkKey) {
		return action.Info{}, errors.New("you can't use jump while on link")
	}

	ai, err := c.Character.Jump(p)

	f := ai.AnimationLength
	ai.Frames = func(action.Action) int { return f }
	ai.AnimationLength = f
	ai.CanQueueAfter = f

	return ai, err
}
