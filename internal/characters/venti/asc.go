package venti

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	a0SwirlKey    = "venti-a0-swirl"
	a0BurstDmgKey = "venti-a0-burst-dmg"
)

// a0HexereiInit registers:
//  1. A permanent AttackMod on Venti giving +35% DmgP to burst attacks while
//     the a0 swirl buff is active.
//  2. Swirl event subscriptions: when burst eye is active and any character
//     triggers Swirl, that character gains +50% DmgP for 4s and Venti gains
//     the a0 swirl buff for 4s.
func (c *char) a0HexereiInit() {
	// A0 requires hexerei mode and 2+ hexerei characters in party
	if !c.isHexerei || !c.hasHexBonus {
		return
	}
	// Permanent burst-attack bonus (+35%) while swirl buff is active
	m := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase(a0BurstDmgKey, -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.ActorIndex != c.Index {
				return nil, false
			}
			if atk.Info.AttackTag != attacks.AttackTagElementalBurst {
				return nil, false
			}
			if !c.StatusIsActive(a0SwirlKey) {
				return nil, false
			}
			for i := range m {
				m[i] = 0
			}
			m[attributes.DmgP] = 0.35
			return m, true
		},
	})

	// Subscribe to all four swirl events
	swirlEvents := []event.Event{
		event.OnSwirlHydro,
		event.OnSwirlPyro,
		event.OnSwirlCryo,
		event.OnSwirlElectro,
	}
	for _, ev := range swirlEvents {
		ev := ev // capture
		c.Core.Events.Subscribe(ev, func(args ...interface{}) bool {
			// Only trigger when Venti's burst eye is active
			if c.Core.F >= c.burstEnd {
				return false
			}
			atk, ok := args[1].(*combat.AttackEvent)
			if !ok {
				return false
			}
			// Apply +50% DmgP AttackMod to the triggering character for 4s (240f)
			chars := c.Core.Player.Chars()
			actorIdx := atk.Info.ActorIndex
			if actorIdx >= 0 && actorIdx < len(chars) {
				triggerBuff := make([]float64, attributes.EndStatType)
				triggerBuff[attributes.DmgP] = 0.50
				chars[actorIdx].AddAttackMod(character.AttackMod{
					Base: modifier.NewBase(
						fmt.Sprintf("venti-a0-swirl-dmg-%d", actorIdx), 240,
					),
					Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
						return triggerBuff, true
					},
				})
			}
			// Activate Venti's a0 swirl status for 4s
			c.AddStatus(a0SwirlKey, 240, false)
			return false
		}, fmt.Sprintf("venti-a0-%v", ev))
	}
}

// A1 is not implemented and will likely never be implemented:
// Holding Skyward Sonnet creates an upcurrent that lasts for 20s.

// Regenerates 15 Energy for Venti after the effects of Wind's Grand Ode end.
// If an Elemental Absorption occurred, this also restores 15 Energy to all characters of that corresponding element in the party.
//
// - checks for ascension level in burst.go to avoid queuing this up only to fail the ascension level check
func (c *char) a4() {
	c.AddEnergy("venti-a4", 15)
	if c.qAbsorb == attributes.NoElement {
		return
	}
	for _, char := range c.Core.Player.Chars() {
		if char.Base.Element == c.qAbsorb {
			char.AddEnergy("venti-a4", 15)
		}
	}
}
