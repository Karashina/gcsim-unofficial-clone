package kachina

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	c.Core.Events.Subscribe(event.OnNightsoulBurst, func(args ...interface{}) bool {
		c.a1buff = make([]float64, attributes.EndStatType)
		c.a1buff[attributes.GeoP] = 0.20
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("kachina-a1", 12*60),
			AffectedStat: attributes.GeoP,
			Amount: func() ([]float64, bool) {
				return c.a1buff, true
			},
		})
		return false
	}, "kachina-a1")
}

func (c *char) a4() {
	if c.Base.Ascension < 4 {
		c.a4flat = 0
		return
	}
	if c.Base.Ascension >= 4 {
		c.a4flat = c.TotalDef() * 0.2
	}
}
