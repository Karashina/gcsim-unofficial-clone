package character

import (
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
)

func (c *CharWrapper) ConsumeNightsoul(amt float64) {
	if !c.HasNightsoul {
		return
	}
	preNightsoul := c.NightsoulPoint
	c.events.Emit(event.OnNightsoulChange, c.Index, -1*amt, "Nightsoul-Drain")
	c.log.NewEvent("draining nightsoul", glog.LogEnergyEvent, c.Index).
		Write("pre_drain", preNightsoul).
		Write("post_drain", c.NightsoulPoint).
		Write("source", c.Base.Key.String()+"-nightsoul-drain").
		Write("max_nightsoul", c.NightsoulPointMax)
	c.NightsoulPoint -= amt
}

func (c *CharWrapper) AddNightsoul(src string, amt float64) {
	if !c.HasNightsoul {
		return
	}

	preNightsoul := c.NightsoulPoint
	c.NightsoulPoint += amt
	if c.NightsoulPoint > c.NightsoulPointMax {
		c.NightsoulPoint = c.NightsoulPointMax
	}
	if c.NightsoulPoint < 0 {
		c.NightsoulPoint = 0
	}

	c.events.Emit(event.OnNightsoulChange, c.Index, amt, src)
	c.log.NewEvent("adding nightsoul", glog.LogEnergyEvent, c.Index).
		Write("rec'd", amt).
		Write("pre_recovery", preNightsoul).
		Write("post_recovery", c.NightsoulPoint).
		Write("source", src).
		Write("max_nightsoul", c.NightsoulPointMax)
}
