package nefer

// C1
// The Base DMG for Lunar-Bloom reactions caused by Nefer's Phantasm Performance is increased by 60% of her Elemental Mastery.
func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}
	// Placeholder for C1 constellation implementation
}

// C2
/*
	Enhances the effects of the A1. Adds duration by 5s, and increases its stack limit to 5. When Nefer unleashes her Elemental Skill, she will instantly gain 2 stacks of Veil of Falsehood.
	Additionally, when Veil of Falsehood reaches 5 stacks, or when the fifth stack's duration is refreshed, Nefer's Elemental Mastery will be increased by 200 for 8s instead of 100 for 8s.
	You must first unlock the A1 to gain this effect.
*/
func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}
	// Placeholder for C2 constellation implementation
}

// C4: Constellation 4
/*
	When Nefer is on the field and in the Shadow Dance state, you will get.
	Additionally, while Nefer is in the Shadow Dance state, nearby opponents will have their Dendro RES decreased by 20%.
	When Nefer exits the Shadow Dance state, this effect will be removed after 4.5s.
*/
func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}
	// add onfield constraint
	c.Core.Player.Verdant.SetGainBonus(c.Core.Player.Verdant.GetGainBonus() + 0.25)
}

// C6: Constellation 6
/*
	When Nefer unleashes Phantasm Performance, the second DMG dealt by herself will be converted to deal AoE Dendro DMG equal to 85% of her Elemental Mastery.
	Additionally, when the attacks from Phantasm Performance end, an extra instance of AoE Dendro DMG equal to 120% of Nefer's Elemental Mastery will be dealt.
	All of the aforementioned DMG is considered Lunar-Bloom DMG dealt by Phantasm Performance.
	TO COPILOT: actual DMG will be implemented in lunarbloomhook.go so just queue dummy attack like ineffa/lauma/flins.

	Moonsign: Ascendant Gleam
	Nefer's Lunar-Bloom DMG is elevated by 15%.
*/
func (c *char) c6() {
	if c.Base.Cons < 6 {
		return
	}
	// Placeholder for C6 constellation implementation
}
