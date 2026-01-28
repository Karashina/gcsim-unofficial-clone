package lauma

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

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
	if c.MoonsignNascent {
		// Moonsign: Nascent Gleam - Bloom, Hyperbloom, and Burgeon DMG can score CRIT Hits
		// CRIT Rate fixed at 15%, CRIT DMG fixed at 100%
		c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
			ae := args[1].(*combat.AttackEvent)

			switch ae.Info.AttackTag {
			case attacks.AttackTagBurningDamage:
			case attacks.AttackTagBloom:
			case attacks.AttackTagHyperbloom:
			case attacks.AttackTagBurgeon:
			default:
				return false
			}

			// Add special crit handling for these reactions
			ae.Snapshot.Stats[attributes.CR] += 0.15
			ae.Snapshot.Stats[attributes.CD] += 1.0

			c.Core.Log.NewEvent("lauma a1 nascent crit buff", glog.LogCharacterEvent, ae.Info.ActorIndex).
				Write("final_crit", ae.Snapshot.Stats[attributes.CR])

			return false
		}, "lauma-a1-nascent-reaction-crit")
	} else if c.MoonsignAscendant {
		c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
			ae := args[1].(*combat.AttackEvent)

			switch ae.Info.AttackTag {
			case attacks.AttackTagLBDamage:
			default:
				return false
			}
			ae.Snapshot.Stats[attributes.CR] += 0.1
			ae.Snapshot.Stats[attributes.CD] += 0.2

			c.Core.Log.NewEvent("lauma a1 nascent crit buff", glog.LogCharacterEvent, ae.Info.ActorIndex)

			return false
		}, "lauma-a1-nascent-reaction-crit")
	}
}

// A4
// Each point of Elemental Mastery Lauma has will give her the following bonuses:
// DMG dealt by her Elemental Skill is increased by 0.04%. The maximum increase obtainable this way is 32%.
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}

	// add Damage Bonus for Elemental Skill
	m := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("lauma-a4-skill-dmg-bonus", -1), // Permanent
		Amount: func(atk *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
			// skip if not skill
			if atk.Info.AttackTag != attacks.AttackTagElementalArt {
				return nil, false
			}
			// calculate EM bonus
			em := c.Stat(attributes.EM)
			bonus := min(0.32, em*0.0004) // 0.04% per EM, max 32%
			m[attributes.DmgP] = bonus
			return m, true
		},
	})
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
