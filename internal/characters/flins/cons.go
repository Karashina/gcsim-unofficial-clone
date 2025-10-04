package flins

const ()

// C1
// The basic cooldown of the special Elemental Skill: Northland Spearstorm is reduced to 4s.
// Additionally, when party members trigger Lunar-Charged reactions, Flins will recover 8 Elemental Energy. This effect can occur once every 5.5s.
func (c *char) c1() {
	if c.Base.Cons < 1 {
		c.northlandCD = 6 * 60
		return
	}
	c.northlandCD = 4 * 60

}

// C2
// For the next 6s after using the special Elemental Skill: Northland Spearstorm, when Flins's next Normal Attack hits an opponent, it will deal an additional 50% of Flins's ATK as AoE Electro DMG. This DMG is considered Lunar-Charged DMG.
// When the moonsign is Moonsign: Ascendant Gleam, While Flins is on the field, after his Electro attacks hit an opponent, that opponent's Electro RES will be decreased by 25% for 7s.
func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}

}

// C4
// Flins's ATK is increased by 20%.
// Additionally, his Ascension Talent "Whispering Flame" is changed: Flins's Elemental Mastery is increased by 10% of his ATK. The maximum increase obtainable this way is 220.
func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}

}

// C6
// The DMG dealt to opponents by Flins's Lunar-Charged reactions is multiplied by 35%.
// When the moonsign is Moonsign: Ascendant Gleam, All nearby party members' Lunar-Charged DMG is multiplied by 10%.
func (c *char) c6() {
	if c.Base.Cons < 6 {
		return
	}

}
