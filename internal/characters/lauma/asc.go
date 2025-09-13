package lauma

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) moonsignInitFunc() {
	count := 0
	for _, char := range c.Core.Player.Chars() {
		if char.StatusIsActive("moonsignKey") {
			count++
		}
	}
	switch count {
	case 1:
		c.moonsignNascent = true // Moonsign: Nascent Gleam
	case 2:
		c.moonsignAscendant = true // Moonsign: Ascendant Gleam
	default:
		c.moonsignNascent = false
		c.moonsignAscendant = false
	}
}

// A0
// Every point of Elemental Mastery that Lauma has increasing Lunar-Bloom's Base DMG by 0.0175%, up to a maximum of 14%.
func (c *char) a0() {
	for _, char := range c.Core.Player.Chars() {
		char.AddStatus("LB-Key", -1, false)
		char.AddLBBaseReactBonusMod(character.LBBaseReactBonusMod{
			Base: modifier.NewBase("Moonsign Benediction: Nature's Chorus (A0)", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				maxval := 0.14
				return min(maxval, c.Stat(attributes.EM)*0.000175), false
			},
		})
	}
}

// A1
// For the next 20s after Lauma uses her Elemental Skill,
// corresponding differing buff effects will be granted depending on the party's Moonsign.
// The buffs provided by different Moonsign levels cannot stack.
// Moonsign: Nascent Gleam
// Bloom, Hyperbloom, and Burgeon DMG dealt by all nearby party members can score CRIT Hits,
// with CRIT Rate fixed at 15%, and CRIT DMG fixed at 100%.
// CRIT Rate from this effect stacks with CRIT Rate from similar effects that allow these Elemental Reactions to CRIT.
// Moonsign: Ascendant Gleam
// All nearby party members' Lunar-Bloom DMG CRIT Rate +10%, CRIT DMG +20%.
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}

}

// A4
// Each point of Elemental Mastery Lauma has will give her the following bonuses:
// DMG dealt by her Elemental Skill is increased by 0.04%. The maximum increase obtainable this way is 32%.
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}

}
