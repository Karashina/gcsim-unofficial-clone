package sigewinne

import (
	"github.com/genshinsim/gcsim/pkg/core/action"
)

func (c *char) Dash(p map[string]int) (action.Info, error) {
	// if burst is active at time of dash
	if c.StatusIsActive(burstkey) {
		c.burstHitSrc = -1       // invalidate any other tasks
		c.DeleteStatus(burstkey) // delete burst
	}

	count := 0
	if p["pickup_droplets"] > 0 {
		count = p["pickup_droplets"]
	}
	c.dropletPickUp(count)

	// call default implementation to handle stamina
	return c.Character.Dash(p)
}
