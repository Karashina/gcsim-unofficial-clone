package lauma

const ()

// C1
// After Lauma uses her Elemental Skill or Elemental Burst, she will gain Threads of Life for 20s.
// During this time, when nearby party members trigger Lunar-Bloom reactions,
// nearby active characters will recover HP equal to 500% of Lauma's Elemental Mastery. This effect can be triggered once every 1.9s.
func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}

}

// C2
// If Moonsign: Ascendant Gleam is active on elemental burst activation, All nearby party members' Lunar-Bloom DMG is increased by 40%.
func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}

}

// C4
// When attacks from the Frostgrove Sanctuary summoned by her Elemental Skill hit opponents,
// Lauma will regain 4 Elemental Energy. This effect can be triggered once every 5s.
func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}
	c.AddEnergy("lauma C4", 4)
}

// C6
// When the Frostgrove Sanctuary attacks opponents, it will deal 1 additional instance of AoE Dendro DMG equal to 185% of Lauma's Elemental Mastery.
// This DMG is considered Lunar-Bloom DMG. This instance of DMG will not consume any Pale Hymn stacks and will provide Lauma with 2 stacks of Pale Hymn,
// as well as refreshing the duration of Pale Hymn stacks gained in this manner.
// This effect can occur up to 8 times during each Frostgrove Sanctuary.
// When using the Elemental Skill Runo: Dawnless Rest of Karsikko, all Pale Hymn stacks gained in this manner will be removed.
// Additionally, when Lauma uses a Normal Attack while she has Pale Hymn stacks,
// she will consume 1 stack to convert this to deal Dendro DMG equal to 150% of her Elemental Mastery. This DMG is considered Lunar-Bloom DMG.
// Moonsign: Ascendant Gleam: All nearby party members' Lunar-Bloom DMG is multiplied by 1.25.
func (c *char) c6() {
	if c.Base.Cons < 6 {
		return
	}

}
