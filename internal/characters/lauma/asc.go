package lauma

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
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
	
	// Apply appropriate Moonsign buffs for 20s
	if c.moonsignNascent {
		// Moonsign: Nascent Gleam
		for _, char := range c.Core.Player.Chars() {
			char.AddReactBonusMod(character.ReactBonusMod{
				Base: modifier.NewBase("lauma-a1-nascent-bloom-crit", 20*60),
				Amount: func(ai combat.AttackInfo) (float64, bool) {
					// For Bloom/Hyperbloom/Burgeon, add crit capability
					if ai.AttackTag == attacks.AttackTagBloom || 
					   ai.AttackTag == attacks.AttackTagHyperbloom || 
					   ai.AttackTag == attacks.AttackTagBurgeon {
						// This would need special reaction crit handling
						return 0, false
					}
					return 0, false
				},
			})
		}
	} else if c.moonsignAscendant {
		// Moonsign: Ascendant Gleam
		for _, char := range c.Core.Player.Chars() {
			char.AddLBReactBonusMod(character.LBReactBonusMod{
				Base: modifier.NewBase("lauma-a1-ascendant-lunar-bloom", 20*60),
				Amount: func(ai combat.AttackInfo) (float64, bool) {
					// Add CRIT Rate +10%, CRIT DMG +20% for Lunar-Bloom
					return 0, false // Would need to modify crit stats
				},
			})
		}
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

// A4 helper for skill damage bonus
func (c *char) a4SkillBonus(ai *combat.AttackInfo) {
	if c.Base.Ascension < 4 {
		return
	}
	em := c.Stat(attributes.EM)
	bonus := min(0.32, em*0.0004) // 0.04% per EM, max 32%
	ai.Mult *= (1 + bonus)
}

// A4 helper for charged attack damage bonus (extending to CA as well)
func (c *char) a4ChargeBonus(ai *combat.AttackInfo) {
	if c.Base.Ascension < 4 {
		return
	}
	em := c.Stat(attributes.EM)
	bonus := min(0.32, em*0.0004) // 0.04% per EM, max 32%
	ai.Mult *= (1 + bonus)
}

// RES reduction from skill hits
// Additionally, when Lauma's Elemental Skill or attacks from Frostgrove Sanctuary hit an opponent,
// that opponent's Dendro RES and Hydro RES will be decreased for 10s.
func (c *char) applyResReduction() {
	// Apply RES reduction through damage event callback
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		if len(args) < 3 {
			return false
		}
		
		enemy, ok := args[0].(combat.Target)
		if !ok {
			return false
		}
		
		atk, ok := args[1].(*combat.AttackEvent)
		if !ok {
			return false
		}
		
		dmg, ok := args[2].(float64)
		if !ok || dmg == 0 {
			return false
		}
		
		// Check if this is from Lauma's skill or sanctuary
		if atk.Info.ActorIndex != c.Index {
			return false
		}
		
		if atk.Info.AttackTag != attacks.AttackTagElementalArt {
			return false
		}
		
		// Apply Dendro and Hydro RES reduction
		if e, ok := enemy.(interface{ AddResistMod(combat.ResistMod) }); ok {
			e.AddResistMod(combat.ResistMod{
				Base:  modifier.NewBaseWithHitlag("lauma-dendro-res-reduction", 10*60),
				Ele:   attributes.Dendro,
				Value: -0.2,
			})
			e.AddResistMod(combat.ResistMod{
				Base:  modifier.NewBaseWithHitlag("lauma-hydro-res-reduction", 10*60),
				Ele:   attributes.Hydro,
				Value: -0.2,
			})
		}
		
		return false
	}, "lauma-res-reduction")
}
