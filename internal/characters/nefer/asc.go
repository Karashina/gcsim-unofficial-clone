package nefer

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// A0
// Every point of Elemental Mastery that Nefer has increases Lunar-Bloom's Base DMG by 0.0175%, up to a maximum of 14%.
func (c *char) a0() {
	for _, char := range c.Core.Player.Chars() {
		char.AddStatus("LB-Key", -1, false)
		char.AddLBBaseReactBonusMod(character.LBBaseReactBonusMod{
			Base: modifier.NewBase("Moonsign Benediction: Dusklit Eaves (A0)", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				maxval := 0.14
				return min(maxval, c.Stat(attributes.EM)*0.000175), false
			},
		})
	}
}

// A1: Seeds/Veil mechanic (simplified): absorbing seeds grants Veil of Falsehood stacks; at thresholds grants EM bonus and increases PP DMG by 8% per stack.
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	// Now handled via direct seed absorption in Charge/Phantasm attacks
}
