package dhalia

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const ()

func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}
	c.AddEnergy("dhalia-c1", 2.5)
}

func (c *char) c6() {
	if c.Base.Cons < 6 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.AtkSpd] = 0.1
	active := c.Core.Player.ActiveChar()
	active.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag("dhalia-c6", c.burstdur),
		AffectedStat: attributes.NoStat,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})
}

