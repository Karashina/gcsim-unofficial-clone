package lauma

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	c1Key = "lauma-c1-threads-of-life"
	c1ICDKey = "lauma-c1-heal-icd"
)

// C1
// After Lauma uses her Elemental Skill or Elemental Burst, she will gain Threads of Life for 20s.
// During this time, when nearby party members trigger Lunar-Bloom reactions,
// nearby active characters will recover HP equal to 500% of Lauma's Elemental Mastery. This effect can be triggered once every 1.9s.
func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}
	
	// Grant Threads of Life buff for 20s
	c.AddStatus(c1Key, 20*60, true)
	
	// Set up healing on Lunar-Bloom reactions
	c.Core.Events.Subscribe(event.OnBloom, func(args ...interface{}) bool {
		if !c.StatusIsActive(c1Key) {
			return false
		}
		if c.StatusIsActive(c1ICDKey) {
			return false
		}
		
		// Check if this is a Lunar-Bloom reaction
		if len(args) < 1 {
			return false
		}
		if attackInfo, ok := args[0].(combat.AttackInfo); ok {
			if attackInfo.AttackTag == attacks.AttackTagLBDamage {
				// Heal all party members
				em := c.Stat(attributes.EM)
				healAmount := 5.0 * em // 500% of EM
				
				for _, char := range c.Core.Player.Chars() {
					char.Heal(&info.HealInfo{
						Caller:  c.Index,
						Target:  char.Index,
						Message: "Threads of Life",
						Src:     healAmount,
						Bonus:   0,
					})
				}
				
				// Set ICD for 1.9s
				c.AddStatus(c1ICDKey, 114, true) // 1.9s * 60 = 114 frames
				return true
			}
		}
		return false
	}, "lauma-c1-heal")
}

// C2
// If Moonsign: Ascendant Gleam is active on elemental burst activation, All nearby party members' Lunar-Bloom DMG is increased by 40%.
func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}
	
	if c.moonsignAscendant {
		// Apply 40% Lunar-Bloom damage bonus for duration of burst effects
		for _, char := range c.Core.Player.Chars() {
			char.AddLBReactBonusMod(character.LBReactBonusMod{
				Base: modifier.NewBase("lauma-c2-ascendant-lb-boost", 30*60), // Same duration as Pale Hymn
				Amount: func(ai combat.AttackInfo) (float64, bool) {
					return 0.4, false // 40% additive bonus
				},
			})
		}
	}
}

// C4
// When attacks from the Frostgrove Sanctuary summoned by her Elemental Skill hit opponents,
// Lauma will regain 4 Elemental Energy. This effect can be triggered once every 5s.
func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}
	if c.StatusIsActive("lauma-c4-energy-icd") {
		return
	}
	c.AddEnergy("lauma C4", 4)
	c.AddStatus("lauma-c4-energy-icd", 5*60, true) // 5s ICD
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
	// Implementation would be complex and require additional tracking
	// This is a placeholder for the C6 effects
}
