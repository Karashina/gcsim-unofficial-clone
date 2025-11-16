package xingqiu

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// A1 is not implemented:
// TODO: When a Rain Sword is shattered or when its duration expires, it regenerates the current character's HP based on 6% of Xingqiu's Max HP.

// Xingqiu gains a 20% Hydro DMG Bonus.
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.HydroP] = 0.2
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("xingqiu-a4", -1),
		AffectedStat: attributes.HydroP,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})
}

