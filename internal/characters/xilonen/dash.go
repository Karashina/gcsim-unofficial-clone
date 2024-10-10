package xilonen

import (
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/glog"
)

func (c *char) Dash(p map[string]int) (action.Info, error) {
	if !c.StatusIsActive(c6buffKey) {
		if c.StatusIsActive(skillKey) {
			if c.Base.Cons < 1 {
				c.ConsumeNightsoul(5)
			} else {
				c.ConsumeNightsoul(3.5)
			}
		} else {
			c.QueueDashStaminaConsumption(p)
		}
	} else {
		c.Core.Log.NewEvent("Dash Stam & Nightsoul consumption Bypassed by C6", glog.LogCharacterEvent, c.Index)
	}

	c.ApplyDashCD()
	length := c.DashLength()
	return action.Info{
		Frames:          func(action.Action) int { return length },
		AnimationLength: length,
		CanQueueAfter:   length,
		State:           action.DashState,
	}, nil
}
