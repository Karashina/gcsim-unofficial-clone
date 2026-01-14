package dhalia

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.AtkSpd] = min((c.MaxHP()/1000)*0.005, 0.2)
	active := c.Core.Player.ActiveChar()
	active.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag("dhalia-a4", c.burstdur),
		AffectedStat: attributes.NoStat,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})
}

