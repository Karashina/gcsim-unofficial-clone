package zibai

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// c1Init initializes Constellation 1: Burst Forth With Vigor, But Enter in Silence
// After using the Elemental Skill, Zibai will immediately gain 100 Phase Shift Radiance,
// and the max number of Spirit Steed's Stride usages per Lunar Phase Shift mode is increased to 5 times.
// Additionally, each time you switch to the Lunar Phase Shift mode, the first Spirit Steed's Stride's
// 2nd-hit Lunar-Crystallize Reaction DMG is increased by 220%.
func (c *char) c1Init() {
	c.maxSpiritSteedUsages = 5

	c.Core.Log.NewEvent("Zibai C1 active: Max Spirit Steed usages increased to 5", glog.LogCharacterEvent, c.Index)
}

// c2Init initializes Constellation 2: At Birth Are Souls Born, and in Death Leave But Husks
// When in the Lunar Phase Shift mode, all nearby party members' Lunar-Crystallize Reaction DMG is increased by 30%.
// When the moonsign is Ascendant Gleam, Ascension Passive A1 is enhanced;
// the DMG dealt by the 2nd hit of Spirit Steed's Stride is further increased by 550% of Zibai's DEF.
// You must first unlock Ascension 1.
func (c *char) c2Init() {
	// Add LCrs reaction bonus for all party members when in Lunar Phase Shift
	for _, char := range c.Core.Player.Chars() {
		char.AddLCrsReactBonusMod(character.LCrsReactBonusMod{
			Base: modifier.NewBase("zibai-c2-lcrs-bonus", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				if !c.lunarPhaseShiftActive {
					return 0, false
				}
				return 0.30, false
			},
		})
	}

	c.Core.Log.NewEvent("Zibai C2 active: Party LCrs bonus +30% during Lunar Phase Shift", glog.LogCharacterEvent, c.Index)
}

// c4Init initializes Constellation 4: The Spirit Passes, Then Form Follows
// While in the Lunar Phase Shift mode, Zibai's Normal Attack sequence will not reset,
// and when Spirit Steed's Stride hits opponents, Zibai will gain the "Scattermoon Splendor" effect:
// The next time she uses Normal Attacks, the additional attack from her 4th hit will deal 250% of
// the original damage as Lunar-Crystallize Reaction DMG.
func (c *char) c4Init() {
	// Scattermoon Splendor is handled in spiritSteedOnHitCB and queueN4AdditionalHit
	// Normal Attack sequence not resetting is handled in attack.go via savedNormalCounter

	c.Core.Log.NewEvent("Zibai C4 active: Scattermoon Splendor and Normal Attack persistence enabled", glog.LogCharacterEvent, c.Index)
}

// c6Init initializes Constellation 6: The World, A Journey in Passing
// While Zibai is in the Lunar Phase Shift mode, her Phase Shift Radiance gain rate is increased by 50%.
// Additionally, Spirit Steed's Stride will change such that it will consume all Phase Shift Radiance.
// This will elevate the DMG dealt by this instance of Spirit Steed's Stride and the Lunar-Crystallize
// Reaction DMG dealt by Zibai within the next 3s by 1.6% for every point consumed above 70.
// This effect cannot stack.
func (c *char) c6Init() {
	// 50% radiance gain increase is handled in addPhaseShiftRadiance
	// Consume all radiance and elevation buff is handled in spiritSteedStride

	c.Core.Log.NewEvent("Zibai C6 active: Enhanced radiance gain and elevation buff", glog.LogCharacterEvent, c.Index)
}

// applyC6ElevationBuff applies the C6 elevation damage buff
func (c *char) applyC6ElevationBuff(bonusPct float64) {
	const c6Duration = 3 * 60 // 3 seconds

	c.AddStatus(c6ElevationBuffKey, c6Duration, true)

	// Add elevation mod for Spirit Steed and LCrs damage
	c.AddElevationMod(character.ElevationMod{
		Base: modifier.NewBaseWithHitlag(c6ElevationBuffKey, c6Duration),
		Amount: func(ai combat.AttackInfo) (float64, bool) {
			// Only apply to Spirit Steed and LCrs attacks from Zibai
			if ai.ActorIndex != c.Index {
				return 0, false
			}
			if ai.AttackTag != attacks.AttackTagElementalArt &&
				ai.AttackTag != attacks.AttackTagLCrsDamage {
				return 0, false
			}
			return bonusPct, false
		},
	})

	c.Core.Log.NewEvent("Zibai C6 Elevation buff applied", glog.LogCharacterEvent, c.Index).
		Write("bonus_pct", bonusPct).
		Write("duration", c6Duration)
}
