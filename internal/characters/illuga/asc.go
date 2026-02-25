package illuga

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	a1CritKey = "illuga-a1-crit"
	a1EmKey   = "illuga-a1-em"
	a1Dur     = 20 * 60 // 20s (I-5 fix: was 12s)
)

// A1: Lightkeeper's Oath
// When Elemental Skill or Burst is used:
// - Other party members' Geo DMG: CRIT Rate +5%, CRIT DMG +10%
// - If Moonsign is Ascendant Gleam, affected party members' EM +50
// Duration: 20s

func (c *char) applyLightkeeperOath() {
	if c.Base.Ascension < 1 {
		return
	}

	// Check if Moonsign is Ascendant Gleam
	isAscendant := c.checkAscendantGleam()

	// Base bonuses
	critRateBonus := 0.05
	critDmgBonus := 0.10
	emBonus := 50.0

	// I-6 fix: C6 CRIT bonuses apply always, EM only when Ascendant
	if c.Base.Cons >= 6 {
		critRateBonus = 0.10 // 10% (vs 5%)
		critDmgBonus = 0.30  // 30% (vs 10%)
	}
	if c.Base.Cons >= 6 && isAscendant {
		emBonus = 80.0 // 80 (vs 50)
	}

	// Apply to all party members (including self for consistency)
	for _, char := range c.Core.Player.Chars() {
		if char.Index == c.Index {
			continue // Don't apply to self, only to other party members
		}

		// Add Geo CRIT bonuses
		m := make([]float64, attributes.EndStatType)
		m[attributes.CR] = critRateBonus
		m[attributes.CD] = critDmgBonus

		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBaseWithHitlag(a1CritKey, a1Dur),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				if atk.Info.Element != attributes.Geo {
					return nil, false
				}
				return m, true
			},
		})

		// If Ascendant Gleam, add EM bonus
		if isAscendant {
			mEM := make([]float64, attributes.EndStatType)
			mEM[attributes.EM] = emBonus

			char.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag(a1EmKey, a1Dur),
				AffectedStat: attributes.EM,
				Amount: func() ([]float64, bool) {
					return mEM, true
				},
			})
		}
	}

	c.Core.Log.NewEvent("Illuga A1: Lightkeeper's Oath applied to party", glog.LogCharacterEvent, c.Index).
		Write("crit_rate_bonus", critRateBonus).
		Write("crit_dmg_bonus", critDmgBonus).
		Write("em_bonus", emBonus).
		Write("is_ascendant", isAscendant)
}

// A4: Enhanced Nightingale's Song
// Nightingale's Song bonuses are enhanced based on party composition:
// Note: Implemented as modifier on Oriole-Song calculations

func (c *char) a4Init() {
	if c.Base.Ascension < 4 {
		return
	}

	// Count Hydro and Geo party members (including self per spec)
	c.a4HydroCount = 0
	c.a4GeoCount = 0

	for _, char := range c.Core.Player.Chars() {
		switch char.Base.Element {
		case attributes.Hydro:
			c.a4HydroCount++
		case attributes.Geo:
			c.a4GeoCount++
		}
	}

	c.Core.Log.NewEvent("Illuga A4: Party composition counted", glog.LogCharacterEvent, c.Index).
		Write("hydro_count", c.a4HydroCount).
		Write("geo_count", c.a4GeoCount)
}

// getA4GeoBonus returns the A4 Nightingale's Song Geo DMG bonus per hit.
// When there are 1/2/3 Hydro or Geo characters in the party,
// increase is equal to 7%/14%/24% of Illuga's EM.
func (c *char) getA4GeoBonus() float64 {
	if c.Base.Ascension < 4 {
		return 0
	}
	count := c.a4HydroCount + c.a4GeoCount
	if count <= 0 {
		return 0
	}
	if count > 3 {
		count = 3
	}
	em := c.Stat(attributes.EM)
	return a4GeoEM[count-1] * em
}

// getA4LCrsBonus returns the A4 Nightingale's Song LCrs DMG bonus per hit.
// When there are 1/2/3 Hydro or Geo characters in the party,
// increase is equal to 48%/96%/160% of Illuga's EM.
func (c *char) getA4LCrsBonus() float64 {
	if c.Base.Ascension < 4 {
		return 0
	}
	count := c.a4HydroCount + c.a4GeoCount
	if count <= 0 {
		return 0
	}
	if count > 3 {
		count = 3
	}
	em := c.Stat(attributes.EM)
	return a4LCrsEM[count-1] * em
}

// checkAscendantGleam checks if the current Moonsign state is Ascendant Gleam
func (c *char) checkAscendantGleam() bool {
	// Check party-wide Moonsign status set during initialization
	return c.MoonsignAscendant
}
