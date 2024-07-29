package sigewinne

import (
	"github.com/genshinsim/gcsim/pkg/core/action"
)

func (c *char) Jump(p map[string]int) (action.Info, error) {
	// if burst is active at time of jump
	if c.StatusIsActive(burstkey) {
		c.burstHitSrc = -1       // invalidate any other tasks
		c.DeleteStatus(burstkey) // delete burst
	}

	return c.Character.Jump(p)
}
