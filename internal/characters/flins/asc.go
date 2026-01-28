package flins

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// A0: Adds base reaction bonus mod for all characters
func (c *char) a0() {
	for _, char := range c.Core.Player.Chars() {
		char.AddStatus("LC-Key", -1, false)
		char.AddLCBaseReactBonusMod(character.LCBaseReactBonusMod{
			Base: modifier.NewBase("Old World Secrets (A0)", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				maxval := 0.14
				return min(maxval, c.TotalAtk()/100*0.007), false
			},
		})
	}
}

// A1
// When the moonsign is Moonsign: Ascendant Gleam, Lunar-Charged reactions triggered by Flins will deal an additional 20% DMG.
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	if c.MoonsignAscendant {
		// Moonsign: Ascendant Gleam
		c.AddLCReactBonusMod(character.LCReactBonusMod{
			Base: modifier.NewBase("Symphony of Winter (A1)", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				return 0.2, false
			},
		})
	}
}

// A4
// Flins's Elemental Mastery is increased by 8% of his ATK. The maximum increase obtainable this way is 160.
// C4 changes this: Flins's Elemental Mastery is increased by 10% of his ATK. The maximum increase obtainable this way is 220.
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag("Whispering Flame (A4)", -1),
		AffectedStat: attributes.EM,
		Extra:        true,
		Amount: func() ([]float64, bool) {
			m := make([]float64, attributes.EndStatType)
			if c.Base.Cons >= 4 {
				// C4: 10% of ATK, max 220
				m[attributes.EM] = min(220, c.TotalAtk()*0.10)
			} else {
				// Base A4: 8% of ATK, max 160
				m[attributes.EM] = min(160, c.TotalAtk()*0.08)
			}
			return m, true
		},
	})
}
