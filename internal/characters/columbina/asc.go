package columbina

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	// A0: Moonsign Benediction
	moonsignKey   = "moonsign"
	lcrsKeyStatus = "lcrs-key"

	// A1: Lunacy
	lunacyKey      = "lunacy"
	lunacyMaxStack = 3
	lunacyDur      = 10 * 60 // 10 seconds
	lunacyCRBonus  = 0.05    // 5% per stack

	// A4: Law of the New Moon
	a4Key     = "law-of-new-moon"
	a4LCDur   = 5 * 60  // 5s cooldown for LC strike
	a4LBDur   = 10 * 60 // 10s cooldown for LB Verdant Dew
	a4LCrsDur = 8 * 60  // 8s cooldown for LCrs attack
)

// A0: Moonsign Benediction
// - Sets "moonsign" and "lcrs-key" status when Columbina is in party
// - Converts Electro-Charged → Lunar-Charged, Bloom → Lunar-Bloom, Hydro Crystallize → Lunar-Crystallize
// - Base DMG Bonus = 0.2% per 1000 Max HP, max 7%
func (c *char) a0Init() {
	// Set Moonsign status on all party members
	for _, char := range c.Core.Player.Chars() {
		char.AddStatus(moonsignKey, -1, false)
		char.AddStatus(lcrsKeyStatus, -1, false)
	}

	// Add base reaction bonus based on Max HP
	// Note: This is a placeholder StatMod for A0 tracking. The actual bonus
	// is applied during Lunar reaction damage calculation (LCBaseReactBonus etc.)
	// We don't add any actual stats here.
}

// a0BaseBonus calculates the base Lunar reaction DMG bonus from Max HP
func (c *char) a0BaseBonus() float64 {
	hp := c.MaxHP()
	bonus := (hp / 1000) * 0.002 // 0.2% per 1000 HP
	if bonus > 0.07 {
		bonus = 0.07 // Max 7%
	}
	return bonus
}

// LCBaseReactBonus returns the base Lunar-Charged reaction bonus
func (c *char) LCBaseReactBonus(ai combat.AttackInfo) float64 {
	return c.a0BaseBonus()
}

// LBBaseReactBonus returns the base Lunar-Bloom reaction bonus
func (c *char) LBBaseReactBonus(ai combat.AttackInfo) float64 {
	return c.a0BaseBonus()
}

// LCrsBaseReactBonus returns the base Lunar-Crystallize reaction bonus
func (c *char) LCrsBaseReactBonus(ai combat.AttackInfo) float64 {
	return c.a0BaseBonus()
}

// LCReactBonus returns additional Lunar-Charged reaction bonus (from Lunar Domain, etc.)
func (c *char) LCReactBonus(ai combat.AttackInfo) float64 {
	bonus := 0.0
	if c.isLunarDomainActive() {
		bonus += burstBonus[c.TalentLvlBurst()]
	}
	return bonus
}

// LBReactBonus returns additional Lunar-Bloom reaction bonus
func (c *char) LBReactBonus(ai combat.AttackInfo) float64 {
	bonus := 0.0
	if c.isLunarDomainActive() {
		bonus += burstBonus[c.TalentLvlBurst()]
	}
	return bonus
}

// LCrsReactBonus returns additional Lunar-Crystallize reaction bonus
func (c *char) LCrsReactBonus(ai combat.AttackInfo) float64 {
	bonus := 0.0
	if c.isLunarDomainActive() {
		bonus += burstBonus[c.TalentLvlBurst()]
	}
	return bonus
}

// A1: Lunacy - CRIT Rate bonus based on stacks
func (c *char) a1Init() {
	if c.Base.Ascension < 1 {
		return
	}

	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase(lunacyKey+"-crit", -1),
		AffectedStat: attributes.CR,
		Amount: func() ([]float64, bool) {
			if c.lunacyStacks <= 0 {
				return nil, false
			}
			m := make([]float64, attributes.EndStatType)
			m[attributes.CR] = float64(c.lunacyStacks) * lunacyCRBonus
			return m, true
		},
	})
}

// a1OnGravityInterference adds Lunacy stack on Gravity Interference trigger
func (c *char) a1OnGravityInterference() {
	if c.Base.Ascension < 1 {
		return
	}

	// Add stack (max 3)
	c.lunacyStacks++
	if c.lunacyStacks > lunacyMaxStack {
		c.lunacyStacks = lunacyMaxStack
	}

	// Refresh duration
	c.AddStatus(lunacyKey, lunacyDur, true)

	c.Core.Log.NewEvent("Lunacy stack gained", glog.LogCharacterEvent, c.Index).
		Write("stacks", c.lunacyStacks).
		Write("crit_bonus", float64(c.lunacyStacks)*lunacyCRBonus)

	// Schedule stack decay
	c.lunacySrc = c.Core.F
	c.Core.Tasks.Add(c.lunacyDecay(c.Core.F), lunacyDur)
}

// lunacyDecay removes Lunacy stacks after duration
func (c *char) lunacyDecay(src int) func() {
	return func() {
		if c.lunacySrc != src {
			return
		}
		c.lunacyStacks = 0
		c.Core.Log.NewEvent("Lunacy expired", glog.LogCharacterEvent, c.Index)
	}
}

// A4: Law of the New Moon - effects based on Lunar reaction type
// When characters within the Lunar Domain trigger Lunar reactions, they gain the following effects
func (c *char) a4Init() {
	if c.Base.Ascension < 4 {
		return
	}

	// Subscribe to EnemyDamage to capture Lunar Charged damage
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		if !c.isLunarDomainActive() {
			return false
		}
		// args: target, AttackEvent, amount, crit
		ae := args[1].(*combat.AttackEvent)
		if ae.Info.AttackTag != attacks.AttackTagLCDamage {
			return false
		}
		c.a4LunarChargedEffect(args...)
		return false
	}, "columbina-a4-lc")

	// Subscribe to all Lunar Bloom events in Lunar Domain
	c.Core.Events.Subscribe(event.OnLunarBloom, func(args ...interface{}) bool {
		if !c.isLunarDomainActive() {
			return false
		}
		c.a4LunarBloomEffect()
		return false
	}, "columbina-a4-lb")

	// Subscribe to all Lunar Crystallize events in Lunar Domain
	c.Core.Events.Subscribe(event.OnLunarCrystallize, func(args ...interface{}) bool {
		if !c.isLunarDomainActive() {
			return false
		}
		c.a4LunarCrystallizeEffect(args...)
		return false
	}, "columbina-a4-lcrs")
}

// a4LunarChargedEffect: On LC - 33% chance to deal same damage as LC
func (c *char) a4LunarChargedEffect(args ...interface{}) {
	if c.Base.Ascension < 4 {
		return
	}

	ae := args[1].(*combat.AttackEvent)
	CRValue := 0
	switch args[3].(bool) {
	case true:
		CRValue = 1
	default:
		CRValue = 0
	}

	// Recursion guard (though checking AttackTagLCDamage in caller helps, strictly speaking A4 shouldn't be LCDamage)
	if ae.Info.Abil == "Law of the New Moon (LC)" {
		return
	}

	// 33% chance to perform additional lightning strike
	if c.Core.Rand.Float64() < 0.33 {
		// Setup A4 attack info mirroring the source LC
		ai := combat.AttackInfo{
			ActorIndex:       c.Index, // Columbina deals the damage
			Abil:             "Law of the New Moon (LC)",
			AttackTag:        attacks.AttackTagElementalBurst, // Use Burst tag or None to avoid triggering A4 loop/modifiers again
			ICDTag:           attacks.ICDTagNone,
			ICDGroup:         attacks.ICDGroupDefault,
			StrikeType:       attacks.StrikeTypeDefault,
			Element:          ae.Info.Element, // Keep source element (Electro/Hydro?)
			Durability:       0,
			IgnoreDefPercent: 1,               // LC ignores def
			FlatDmg:          ae.Info.FlatDmg, // Copy calculated damage
		}

		// Use empty snapshot to avoid double-applying stats
		// The damage specific to the reaction is already in FlatDmg
		snap := combat.Snapshot{
			// We might need CharLvl for defense calculation if IgnoreDef triggers based on diff?
			// But IgnoreDefPercent is 1.
			CharLvl: c.Base.Level,
		}
		// Borrow CR from original hit CD is zeroed to avoid double crit damage
		snap.Stats[attributes.CR] = float64(CRValue)
		snap.Stats[attributes.CD] = 0

		c.Core.Tasks.Add(func() {
			ap := combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget().Pos(), nil, 8)
			c.Core.QueueAttackWithSnap(ai, snap, ap, 0)
			c.Core.Log.NewEvent("A4 LC additional strike", glog.LogCharacterEvent, c.Index).
				Write("mirrored_flat_dmg", ai.FlatDmg)
		}, 10)
	}
}

// a4LunarBloomEffect: On LB - gain 1 Moonridge Dew every 18s (max 3)
func (c *char) a4LunarBloomEffect() {
	if c.Base.Ascension < 4 {
		return
	}

	// Check if on cooldown (18s from last gain)
	if c.Core.F < c.moonridgeICD {
		return
	}

	// Gain 1 Moonridge Dew
	if c.moonridgeDew < 3 {
		c.moonridgeDew++
		c.moonridgeICD = c.Core.F + 18*60 // 18s cooldown

		c.Core.Log.NewEvent("A4 Moonridge Dew gained", glog.LogCharacterEvent, c.Index).
			Write("dew", c.moonridgeDew)
	}
}

// a4LunarCrystallizeEffect: On LCrs - 33% chance for Moondrift additional attack
func (c *char) a4LunarCrystallizeEffect(args ...interface{}) {
	if c.Base.Ascension < 4 {
		return
	}

	// Prevent chain reaction from self
	if len(args) > 1 {
		ae := args[1].(*combat.AttackEvent)
		if ae.Info.Abil == "Law of the New Moon (Geo)" {
			return
		}
	}

	// 33% chance for each Moondrift to inflict an extra attack
	if c.Core.Rand.Float64() < 0.33 {
		c.Core.Tasks.Add(func() {
			ai := combat.AttackInfo{
				ActorIndex:       c.Index,
				Abil:             "Law of the New Moon (Geo)",
				AttackTag:        attacks.AttackTagLCrsDamage,
				ICDTag:           attacks.ICDTagNone,
				ICDGroup:         attacks.ICDGroupDefault,
				StrikeType:       attacks.StrikeTypeDefault,
				Element:          attributes.Geo,
				Durability:       0,
				IgnoreDefPercent: 1,
			}

			em := c.Stat(attributes.EM)
			// Use Gravity Interference scaling as a baseline
			baseDmg := c.MaxHP() * gravityInterfLCrs[c.TalentLvlSkill()] * (1 + c.LCrsBaseReactBonus(ai))
			emBonus := (16 * em) / (2000 + em)
			ai.FlatDmg = baseDmg * (1 + emBonus + c.LCrsReactBonus(ai)) * (1 + c.ElevationBonus(ai))

			snap := combat.Snapshot{CharLvl: c.Base.Level}
			snap.Stats[attributes.CR] = c.Stat(attributes.CR)
			snap.Stats[attributes.CD] = c.Stat(attributes.CD)

			ap := combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget().Pos(), nil, 8)
			c.Core.QueueAttackWithSnap(ai, snap, ap, 60) // 1s delay

			c.Core.Log.NewEvent("A4 LCrs additional Moondrift attack", glog.LogCharacterEvent, c.Index)
		}, 10)
	}
}
