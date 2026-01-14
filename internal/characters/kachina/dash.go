package kachina

import (
	"errors"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
)

func (c *char) Dash(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(skillRideKey) {
		return action.Info{}, errors.New("you can't use dash while riding Turbo Twirly")
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

