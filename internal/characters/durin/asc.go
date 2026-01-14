package durin

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/reactions"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	// A1 state keys
	a1ResShredKey  = "durin-a1-res-shred"
	a1DarkDecayKey = "durin-a1-dark-decay"

	// A1 durations and values
	a1ResShredDuration = 6 * 60 // 6 seconds

	// A4 state keys
	a4PrimordialKey = "durin-a4-primordial"

	// A4 durations and values
	a4Duration         = 20 * 60 // 20 seconds
	a4MaxStacks        = 10
	a4DmgPercentPerAtk = 0.03 // 3% per 100 ATK
	a4MaxDmgPercent    = 0.75 // Maximum 75% additional DMG
)

// A1: Light Manifest of the Divine Calculus
// Dragon of White Flame: After Burning, Overloaded, Pyro Swirl, or Pyro Crystallize reactions,
// or dealing Pyro/Dendro DMG to burning enemies, decrease Pyro RES and corresponding elemental RES by 20%
// Dragon of Dark Decay: Durin's Vaporize and Melt DMG increased by 40%
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}

	// Subscribe to reaction events for Dragon of White Flame RES shred
	c.Core.Events.Subscribe(event.OnBurning, c.a1WhiteFlameReactionCB(attributes.Pyro, attributes.Dendro), "durin-a1-burning")
	c.Core.Events.Subscribe(event.OnOverload, c.a1WhiteFlameReactionCB(attributes.Pyro, attributes.Electro), "durin-a1-overload")
	c.Core.Events.Subscribe(event.OnSwirlPyro, c.a1WhiteFlameReactionCB(attributes.Pyro, attributes.Anemo), "durin-a1-pyro-swirl")
	c.Core.Events.Subscribe(event.OnCrystallizePyro, c.a1WhiteFlameReactionCB(attributes.Pyro, attributes.Geo), "durin-a1-pyro-crystallize")

	// Subscribe for Pyro/Dendro DMG to burning enemies
	c.Core.Events.Subscribe(event.OnEnemyDamage, c.a1WhiteFlameOnDamageCB, "durin-a1-burning-dmg")

	// Dragon of Dark Decay: reactmod for Vaporize and Melt
	c.a1DarkDecayReactMod()
}

func (c *char) a1WhiteFlameReactionCB(ele1, ele2 attributes.Element) func(args ...interface{}) bool {
	return func(args ...interface{}) bool {
		if !c.StatusIsActive(dragonWhiteFlameKey) {
			return false
		}

		target := args[0].(combat.Target)
		e, ok := target.(*enemy.Enemy)
		if !ok {
			return false
		}

		// Calculate RES shred amount (20% or 35% with Hexerei bonus)
		resShred := 0.20
		if c.hasHexereiBonus() {
			resShred = 0.35 // 20% * 1.75
		}

		// Apply RES shred for both elements
		e.AddResistMod(combat.ResistMod{
			Base:  modifier.NewBaseWithHitlag(a1ResShredKey+"-"+ele1.String(), a1ResShredDuration),
			Ele:   ele1,
			Value: -resShred,
		})
		if ele1 != ele2 {
			e.AddResistMod(combat.ResistMod{
				Base:  modifier.NewBaseWithHitlag(a1ResShredKey+"-"+ele2.String(), a1ResShredDuration),
				Ele:   ele2,
				Value: -resShred,
			})
		}

		c.Core.Log.NewEvent("Durin A1 RES shred applied", glog.LogCharacterEvent, c.Index).
			Write("element1", ele1.String()).
			Write("element2", ele2.String()).
			Write("shred", resShred)

		return false
	}
}

func (c *char) a1WhiteFlameOnDamageCB(args ...interface{}) bool {
	if !c.StatusIsActive(dragonWhiteFlameKey) {
		return false
	}

	atk := args[1].(*combat.AttackEvent)
	target := args[0].(combat.Target)
	e, ok := target.(*enemy.Enemy)
	if !ok {
		return false
	}

	// Check if target is burning
	if !e.AuraContains(attributes.Pyro, attributes.Dendro) {
		return false
	}

	// Check if dealing Pyro or Dendro DMG
	if atk.Info.Element != attributes.Pyro && atk.Info.Element != attributes.Dendro {
		return false
	}

	// Calculate RES shred amount
	resShred := 0.20
	if c.hasHexereiBonus() {
		resShred = 0.35
	}

	// Apply Pyro RES shred and corresponding element RES shred
	e.AddResistMod(combat.ResistMod{
		Base:  modifier.NewBaseWithHitlag(a1ResShredKey+"-pyro", a1ResShredDuration),
		Ele:   attributes.Pyro,
		Value: -resShred,
	})
	e.AddResistMod(combat.ResistMod{
		Base:  modifier.NewBaseWithHitlag(a1ResShredKey+"-"+atk.Info.Element.String(), a1ResShredDuration),
		Ele:   atk.Info.Element,
		Value: -resShred,
	})

	return false
}

func (c *char) a1DarkDecayReactMod() {
	// Apply reactmod for Vaporize and Melt when Dragon of Dark Decay is active
	reactMod := 0.40
	if c.hasHexereiBonus() {
		reactMod = 0.70 // 40% * 1.75
	}

	c.AddReactBonusMod(character.ReactBonusMod{
		Base: modifier.NewBase(a1DarkDecayKey, -1),
		Amount: func(ai combat.AttackInfo) (float64, bool) {
			if !c.StatusIsActive(dragonDarkDecayKey) {
				return 0, false
			}
			if ai.ActorIndex != c.Index {
				return 0, false
			}
			if !ai.Amped {
				return 0, false
			}
			return reactMod, true
		},
	})
}

func (c *char) a4OnBurst() {
	if c.Base.Ascension < 4 {
		return
	}

	// Reset stacks when burst is used
	c.primordialFusionStacks = a4MaxStacks
	c.primordialFusionExpiry = c.Core.F + a4Duration
	c.AddStatus(a4PrimordialKey, a4Duration, true)

	c.Core.Log.NewEvent("Durin A4: Primordial Fusion stacks gained", glog.LogCharacterEvent, c.Index).
		Write("stacks", c.primordialFusionStacks)
}

func (c *char) a4DragonAttackCB(a combat.AttackCB) {
	if c.Base.Ascension < 4 {
		return
	}

	// Check if we have stacks
	if c.primordialFusionStacks <= 0 || c.Core.F >= c.primordialFusionExpiry {
		return
	}

	// Consume 1 stack (only once per attack, even if hitting multiple enemies)
	c.primordialFusionStacks--

	// Calculate DMG bonus: 3% per 100 ATK, max 75%
	// The spec says to multiply ai.Mult, but since we're in a callback,
	// we need to apply this as flat damage or similar mechanism
	totalAtk := c.TotalAtk()
	dmgBonus := (totalAtk / 100.0) * a4DmgPercentPerAtk
	if dmgBonus > a4MaxDmgPercent {
		dmgBonus = a4MaxDmgPercent
	}

	c.Core.Log.NewEvent("Durin A4: Primordial Fusion consumed", glog.LogCharacterEvent, c.Index).
		Write("stacks_remaining", c.primordialFusionStacks).
		Write("atk", totalAtk).
		Write("dmg_bonus", dmgBonus)

	// Note: The actual damage modification is handled in the attack info setup
	// since we can't modify damage in a callback. This callback is for stack consumption.
}

// Vaporize and Melt reaction types for A1 Dark Decay check
func init() {
	_ = reactions.Vaporize
	_ = reactions.Melt
}
