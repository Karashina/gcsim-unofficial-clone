package aino

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
)

// A1 - Modular Efficiency Protocol
// When the Moonsign is Ascendant, Her Elemental Burst Precision Hydronic Cooler is enhanced:
// The Cool Your Jets Ducky will fire water balls more frequently, and the water balls will deal AoE Hydro DMG over a larger area of effect.
// This is handled in burst.go

func (c *char) a1() {
	// Implementation is in burst.go where interval and radius are determined
}

// A4 - Structured Power Booster
// Aino's Elemental Burst FlatDMG is increased by 50% of her Elemental Mastery.
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}

	// Update flat damage buff whenever EM changes
	c.AddStatMod("aino-a4-em-buff", -1, attributes.NoStat, func() ([]float64, bool) {
		em := c.Stat(attributes.EM)
		c.a4FlatDmgBuff = 0.5 * em
		return nil, false
	})
}
