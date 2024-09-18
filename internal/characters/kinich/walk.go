package kinich

import (
	"github.com/genshinsim/gcsim/pkg/core/action"
)

func (c *char) Walk(p map[string]int) (action.Info, error) {
	if p["blindspot"] > 0 && c.StatusIsActive(skillKey) && c.StatusIsActive(blindspotKey) {
		c.AddNightsoul("kinich-blindspot", 4)
		c.DeleteStatus(blindspotKey)
	}
	// call default implementation to handle stamina
	return c.Character.Walk(p)
}
