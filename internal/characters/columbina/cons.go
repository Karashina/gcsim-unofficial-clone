package columbina

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/shield"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	// C1
	c1Key            = "c1-gravity-interference"
	c1GravitySkipKey = "c1-gravity-skip"
	c1ICD            = 15 * 60 // 15s ICD
	c1Elevation      = 0.015   // 1.5% Elevation bonus for all party members

	// C2
	c2Key = "lunar-brilliance"
	c2Dur = 8 * 60 // 8s duration

	// C4
	c4IcdKey      = "c4-gravity-bonus-icd"
	c4Energy      = 4
	c4HPBonusLC   = 0.125 // 12.5% Max HP
	c4HPBonusLB   = 0.025 // 2.5% Max HP
	c4HPBonusLCrs = 0.125 // 12.5% Max HP

	// C6
	c6Key       = "c6-crit-dmg"
	c6Dur       = 8 * 60 // 8s duration
	c6CDBonus   = 0.80   // 80% CRIT DMG
	c6Elevation = 0.07   // 7% Elevation
)

// C1: On skill cast, trigger Gravity Interference effect (once per 15s)
// Effect provides energy restoration, poise restoration, or shield based on dominant type
// Also provides 1.5% Elevation bonus to all party members' Lunar Reaction DMG
func (c *char) c1Init() {
	if c.Base.Cons < 1 {
		return
	}

	// Apply 1.5% Elevation bonus to all party members for Lunar Reaction DMG only
	// This is evaluated during precalc (calcLunarChargedDmg, etc.) where atk has the correct AttackTag
	for _, char := range c.Core.Player.Chars() {
		char.AddElevationMod(character.ElevationMod{
			Base: modifier.NewBase("columbina-c1-elevation", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				if ai.AttackTag == attacks.AttackTagLCDamage ||
					ai.AttackTag == attacks.AttackTagLBDamage ||
					ai.AttackTag == attacks.AttackTagLCrsDamage {
					return c1Elevation, false
				}
				return 0, false
			},
		})
	}
}

func (c *char) c1OnSkill() {
	if c.Base.Cons < 1 {
		return
	}
	if c.StatusIsActive(c1Key) {
		return
	}
	c.AddStatus(c1Key, c1ICD, false)

	dominantType := c.getDominantLunarType()

	c.Core.Log.NewEvent("C1 Gravity Interference triggered on skill", glog.LogCharacterEvent, c.Index).
		Write("dominant_type", dominantType)

	c.AddStatus(c1GravitySkipKey, -1, false)
	c.triggerGravityInterference()

}

func (c *char) c1OnGravityInterference() {
	if c.Base.Cons < 1 {
		return
	}

	dominantType := c.getDominantLunarType()

	switch dominantType {
	case "lc":
		// Energy restoration
		c.AddEnergy("c1-energy", 6)
	case "lcrs":
		// Summon Rainsea Shield: 12% Max HP, 250% effectiveness vs Hydro DMG, 8s duration
		shieldAmount := c.MaxHP() * 0.12
		// Rainsea Shield implementation
		// 英語コメント: Apply Rainsea Shield (Hydro 250% effectiveness, 8s duration)
		importShield := func() {
			// gcsim標準のシールドAPIを使用
			// ShieldType: use EndType+1 for custom (or define ColumbinaShield in shield.go if needed)
			// Target: active character
			c.Core.Player.Shields.Add(&shield.Tmpl{
				ActorIndex: c.Index,
				Target:     c.Core.Player.Active(),
				Name:       "Rainsea Shield",
				Src:        c.Index,
				ShieldType: shield.ColumbinaC1,
				Ele:        attributes.Hydro,
				HP:         shieldAmount,
				Expires:    c.Core.F + 8*60,
			})
			c.Core.Log.NewEvent("C1 Rainsea Shield applied", glog.LogCharacterEvent, c.Index).
				Write("amount", shieldAmount)
		}
		importShield()
	}
}

// C2: Rate of accumulating Gravity increases by 34%
// On Gravity Interference, gain Lunar Brilliance (40% Max HP for 8s) with stat buffs based on dominant Lunar type
// When moonsign is Ascendant Gleam, apply buffs to active character based on dominant type:
// - LC: ATK = 1% of Columbina's Max HP
// - LB: EM = 0.35% of Columbina's Max HP
// - LCrs: DEF = 1% of Columbina's Max HP
func (c *char) c2OnGravityInterference(dominantType string) {
	if c.Base.Cons < 2 {
		return
	}

	// Gain Lunar Brilliance (40% Max HP boost to Columbina for 8s)
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag(c2Key+"-hp", c2Dur),
		AffectedStat: attributes.HPP,
		Amount: func() ([]float64, bool) {
			m := make([]float64, attributes.EndStatType)
			m[attributes.HPP] = 0.40 // 40% Max HP
			return m, true
		},
	})

	// If Moonsign is Ascendant Gleam, apply additional buffs to active character
	if !c.MoonsignAscendant {
		return
	}

	columbinaMHP := c.Stat(attributes.HP)
	activeCharIdx := c.Core.Player.Active()
	activeChar := c.Core.Player.ByIndex(activeCharIdx)

	// Lunar Brilliance buffs based on dominant type
	switch dominantType {
	case "lc":
		// ATK increase = 1% of Columbina's Max HP
		buffValue := 0.01 * columbinaMHP
		activeChar.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(c2Key+"-atk", c2Dur),
			AffectedStat: attributes.ATK,
			Amount: func() ([]float64, bool) {
				m := make([]float64, attributes.EndStatType)
				m[attributes.ATK] = buffValue
				return m, true
			},
		})
	case "lb":
		// EM increase = 0.35% of Columbina's Max HP
		buffValue := 0.0035 * columbinaMHP
		activeChar.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(c2Key+"-em", c2Dur),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				m := make([]float64, attributes.EndStatType)
				m[attributes.EM] = buffValue
				return m, true
			},
		})
	case "lcrs":
		// DEF increase = 1% of Columbina's Max HP
		buffValue := 0.01 * columbinaMHP
		activeChar.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(c2Key+"-def", c2Dur),
			AffectedStat: attributes.DEF,
			Amount: func() ([]float64, bool) {
				m := make([]float64, attributes.EndStatType)
				m[attributes.DEF] = buffValue
				return m, true
			},
		})
	}

	c.Core.Log.NewEvent("C2 Lunar Brilliance activated", glog.LogCharacterEvent, c.Index).
		Write("dominant_type", dominantType).
		Write("columbina_mhp", columbinaMHP).
		Write("duration", c2Dur)
}

// C4: On Gravity Interference, restore 4 Energy and add HP scaling DMG bonus
// Lunar Reaction DMG is increased by 12.5%/2.5%/12.5% of Max HP for LC/LB/LCrs respectively
func (c *char) c4OnGravityInterference(dominantType string) {
	if c.Base.Cons < 4 {
		return
	}

	// Restore energy
	c.AddEnergy("c4-energy", c4Energy)

	// Record dominant type for C4 bonus application
	c.c4DominantType = dominantType

	c.Core.Log.NewEvent("C4 energy restored", glog.LogCharacterEvent, c.Index).
		Write("energy", c4Energy).
		Write("dominant_type", dominantType)
}

// C6: For 8s after characters in the Lunar Domain trigger a Lunar reaction,
// based on the elements involved, the CRIT DMG for corresponding Elemental DMG is increased by 80%.
// Effects for the same Element do not stack.
// Also provides 7% Elevation bonus to all party members' Lunar Reaction DMG.
func (c *char) c6Init() {
	if c.Base.Cons < 6 {
		return
	}

	// Apply 7% Elevation bonus to all party members for Lunar Reaction DMG
	for _, char := range c.Core.Player.Chars() {
		char.AddElevationMod(character.ElevationMod{
			Base: modifier.NewBase("columbina-c6-elevation", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				if ai.AttackTag == attacks.AttackTagLCDamage ||
					ai.AttackTag == attacks.AttackTagLBDamage ||
					ai.AttackTag == attacks.AttackTagLCrsDamage {
					return c6Elevation, false
				}
				return 0, false
			},
		})
	}

	// Subscribe to all Lunar reaction events to apply CRIT DMG buff
	c.Core.Events.Subscribe(event.OnLunarCharged, func(args ...interface{}) bool {
		c.c6ApplyBuffToAllChars(attributes.Electro)
		c.c6ApplyBuffToAllChars(attributes.Hydro)
		return false
	}, "columbina-c6-lc-trigger")

	c.Core.Events.Subscribe(event.OnLunarBloom, func(args ...interface{}) bool {
		c.c6ApplyBuffToAllChars(attributes.Dendro)
		c.c6ApplyBuffToAllChars(attributes.Hydro)
		return false
	}, "columbina-c6-lb-trigger")

	c.Core.Events.Subscribe(event.OnLunarCrystallize, func(args ...interface{}) bool {
		c.c6ApplyBuffToAllChars(attributes.Geo)
		c.c6ApplyBuffToAllChars(attributes.Hydro)
		return false
	}, "columbina-c6-lcrs-trigger")
}

// c6ApplyBuffToAllChars applies C6 CRIT DMG buff to all party members for corresponding element
func (c *char) c6ApplyBuffToAllChars(element attributes.Element) {
	if c.Base.Cons < 6 {
		return
	}

	// Apply CRIT DMG buff to all party members
	for _, char := range c.Core.Player.Chars() {
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(c6Key+"-"+element.String(), c6Dur),
			AffectedStat: attributes.CD,
			Amount: func() ([]float64, bool) {
				m := make([]float64, attributes.EndStatType)
				m[attributes.CD] = c6CDBonus
				return m, true
			},
		})
	}

	c.Core.Log.NewEvent("C6 CRIT DMG buff applied to all party members", glog.LogCharacterEvent, c.Index).
		Write("element", element.String()).
		Write("crit_dmg_bonus", c6CDBonus).
		Write("duration", c6Dur)
}

// C3: All nearby party members' Lunar Reaction DMG is elevated by 1.5%
// C5: All nearby party members' Lunar Reaction DMG is elevated by 1.5%
// These stack with C1's 1.5% for a total of 4.5% at C5
func (c *char) c3c5Init() {
	// C3: +1.5% Elevation
	if c.Base.Cons >= 3 {
		for _, char := range c.Core.Player.Chars() {
			char.AddElevationMod(character.ElevationMod{
				Base: modifier.NewBase("columbina-c3-elevation", -1),
				Amount: func(ai combat.AttackInfo) (float64, bool) {
					if ai.AttackTag == attacks.AttackTagLCDamage ||
						ai.AttackTag == attacks.AttackTagLBDamage ||
						ai.AttackTag == attacks.AttackTagLCrsDamage {
						return 0.015, false
					}
					return 0, false
				},
			})
		}
	}

	// C5: +1.5% Elevation (stacks with C3)
	if c.Base.Cons >= 5 {
		for _, char := range c.Core.Player.Chars() {
			char.AddElevationMod(character.ElevationMod{
				Base: modifier.NewBase("columbina-c5-elevation", -1),
				Amount: func(ai combat.AttackInfo) (float64, bool) {
					if ai.AttackTag == attacks.AttackTagLCDamage ||
						ai.AttackTag == attacks.AttackTagLBDamage ||
						ai.AttackTag == attacks.AttackTagLCrsDamage {
						return 0.015, false
					}
					return 0, false
				},
			})
		}
	}
}
