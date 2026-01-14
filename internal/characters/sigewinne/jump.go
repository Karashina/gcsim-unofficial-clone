package sigewinne

import "github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"

func (c *char) Jump(p map[string]int) (action.Info, error) {
	c.burstEarlyCancelled = false
	return c.Character.Jump(p)
}

