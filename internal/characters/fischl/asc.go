package fischl

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const a4IcdKey = "fischl-a4-icd"

// Hexerei Passive constants
const (
	hexOverloadAtkKey = "fischl-hex-overload-atk"
	hexECEMKey        = "fischl-hex-ec-em"
	hexC6BoostKey     = "fischl-hex-c6-boost"
	hexDuration       = 10 * 60 // 10 seconds
	hexOverloadAtkPct = 0.225   // 22.5% ATK
	hexECEM           = 90      // +90 EM
	hexC6Multiplier   = 2.0     // 2x boost when C6 hits
)

// A1 is not implemented:
// TODO: When Fischl hits Oz with a fully-charged Aimed Shot, Oz brings down Thundering Retribution, dealing AoE Electro DMG equal to 152.7% of the arrow's DMG.

// If your current active character triggers an Electro-related Elemental Reaction when Oz is on the field,
// the opponent shall be stricken with Thundering Retribution that deals Electro DMG equal to 80% of Fischl's ATK.
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}

	// Hyperbloom comes from a gadget so it doesn't ignore gadgets
	//nolint:unparam // ignoring for now, event refactor should get rid of bool return of event sub
	a4cb := func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)

		if ae.Info.ActorIndex != c.Core.Player.Active() {
			return false
		}
		// do nothing if oz not on field
		if !c.StatusIsActive(ozActiveKey) {
			return false
		}
		active := c.Core.Player.ActiveChar()
		if active.StatusIsActive(a4IcdKey) {
			return false
		}
		active.AddStatus(a4IcdKey, 0.5*60, true)

		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Fischl A4",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupFischl,
			StrikeType: attacks.StrikeTypePierce,
			Element:    attributes.Electro,
			Durability: 25,
			Mult:       0.8,
		}

		// A4 uses Oz Snapshot
		// TODO: this should target closest enemy within 15m of "elemental reaction position"
		c.Core.QueueAttackWithSnap(
			ai,
			c.ozSnapshot.Snapshot,
			combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 0.5),
			4)
		return false
	}

	a4cbNoGadget := func(args ...interface{}) bool {
		if _, ok := args[0].(*enemy.Enemy); !ok {
			return false
		}
		return a4cb(args...)
	}

	c.Core.Events.Subscribe(event.OnOverload, a4cbNoGadget, "fischl-a4")
	c.Core.Events.Subscribe(event.OnElectroCharged, a4cbNoGadget, "fischl-a4")
	c.Core.Events.Subscribe(event.OnSuperconduct, a4cbNoGadget, "fischl-a4")
	c.Core.Events.Subscribe(event.OnSwirlElectro, a4cbNoGadget, "fischl-a4")
	c.Core.Events.Subscribe(event.OnCrystallizeElectro, a4cbNoGadget, "fischl-a4")
	c.Core.Events.Subscribe(event.OnHyperbloom, a4cb, "fischl-a4")
	c.Core.Events.Subscribe(event.OnQuicken, a4cbNoGadget, "fischl-a4")
	c.Core.Events.Subscribe(event.OnAggravate, a4cbNoGadget, "fischl-a4")
}

// Hexerei Passive:
// When Oz is on the field, team characters gain the following buffs:
// - After triggering Overload: Fischl and active character gain +22.5% ATK for 10s
// - After triggering Electro-Charged or Aggravate: Fischl and active character gain +90 EM for 10s
// - When C6 is unlocked and C6 attack hits: Above effects are doubled for 10s
func (c *char) hexPassive() {
	if !c.isHexerei {
		return
	}

	// Helper to apply ATK% buff
	applyAtkBuff := func(multiplier float64) {
		atkPct := hexOverloadAtkPct * multiplier

		// Buff Fischl
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(hexOverloadAtkKey, hexDuration),
			AffectedStat: attributes.ATKP,
			Amount: func() ([]float64, bool) {
				if !c.StatusIsActive(ozActiveKey) {
					return nil, false
				}
				var stats attributes.Stats
				stats[attributes.ATKP] = atkPct
				return stats[:], true
			},
		})

		// Buff active character
		active := c.Core.Player.ActiveChar()
		active.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(hexOverloadAtkKey, hexDuration),
			AffectedStat: attributes.ATKP,
			Amount: func() ([]float64, bool) {
				if !c.StatusIsActive(ozActiveKey) {
					return nil, false
				}
				var stats attributes.Stats
				stats[attributes.ATKP] = atkPct
				return stats[:], true
			},
		})
	}

	// Helper to apply EM buff
	applyEMBuff := func(multiplier float64) {
		em := hexECEM * multiplier

		// Buff Fischl
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(hexECEMKey, hexDuration),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				if !c.StatusIsActive(ozActiveKey) {
					return nil, false
				}
				var stats attributes.Stats
				stats[attributes.EM] = em
				return stats[:], true
			},
		})

		// Buff active character
		active := c.Core.Player.ActiveChar()
		active.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(hexECEMKey, hexDuration),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				if !c.StatusIsActive(ozActiveKey) {
					return nil, false
				}
				var stats attributes.Stats
				stats[attributes.EM] = em
				return stats[:], true
			},
		})
	}

	// Subscribe to Overload
	c.Core.Events.Subscribe(event.OnOverload, func(args ...interface{}) bool {
		// Check if Oz is active
		if !c.StatusIsActive(ozActiveKey) {
			return false
		}

		// Determine multiplier based on C6 boost status
		multiplier := 1.0
		if c.StatusIsActive(hexC6BoostKey) {
			multiplier = hexC6Multiplier
		}

		applyAtkBuff(multiplier)
		return false
	}, "fischl-hex-overload")

	// Subscribe to Electro-Charged and Aggravate
	ecCallback := func(args ...interface{}) bool {
		// Check if Oz is active
		if !c.StatusIsActive(ozActiveKey) {
			return false
		}

		// Determine multiplier based on C6 boost status
		multiplier := 1.0
		if c.StatusIsActive(hexC6BoostKey) {
			multiplier = hexC6Multiplier
		}

		applyEMBuff(multiplier)
		return false
	}

	c.Core.Events.Subscribe(event.OnElectroCharged, ecCallback, "fischl-hex-ec")
	c.Core.Events.Subscribe(event.OnAggravate, ecCallback, "fischl-hex-aggravate")
}

// C6 Hexerei boost: Subscribe to OnEnemyHit to detect C6 attacks
func (c *char) hexC6Boost() {
	if !c.isHexerei || c.Base.Cons < 6 {
		return
	}

	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)

		// Check if this is a C6 attack from Fischl
		if ae.Info.ActorIndex != c.Index {
			return false
		}
		if ae.Info.Abil != "Fischl C6" {
			return false
		}

		// Activate C6 boost for 10 seconds
		c.AddStatus(hexC6BoostKey, hexDuration, true)
		return false
	}, "fischl-hex-c6-boost")
}
