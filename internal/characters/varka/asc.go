package varka

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// a1Init sets up the A1 passive: per 1000 ATK → 10% Anemo + Other DMG Bonus (max 25%)
// Uses AttackMod to avoid infinite recursion from TotalAtk() inside StatMod
func (c *char) a1Init() {
	if c.Base.Ascension < 1 {
		return
	}
	otherP := attributes.NoStat
	if c.hasOtherEle {
		otherP = eleToStatP(c.otherElement)
	}

	m := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase(a1Key, -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			totalAtk := c.TotalAtk()
			bonus := totalAtk / 1000.0 * 0.10
			if bonus > 0.25 {
				bonus = 0.25
			}
			// Reset
			for i := range m {
				m[i] = 0
			}
			m[attributes.AnemoP] = bonus
			if otherP != attributes.NoStat {
				m[otherP] = bonus
			}
			return m, true
		},
	})
}

// a4Init subscribes to Swirl events for Azure Fang's Oath stacks
func (c *char) a4Init() {
	swirlEvents := []event.Event{
		event.OnSwirlHydro,
		event.OnSwirlPyro,
		event.OnSwirlCryo,
		event.OnSwirlElectro,
	}
	for _, ev := range swirlEvents {
		c.Core.Events.Subscribe(ev, func(args ...interface{}) bool {
			atk := args[1].(*combat.AttackEvent)
			// ICD: each character can grant 1 stack per second
			charIdx := atk.Info.ActorIndex
			icdKey := fmt.Sprintf("%s%d", a4ICDPrefix, charIdx)
			if c.StatusIsActive(icdKey) {
				return false
			}
			c.AddStatus(icdKey, 60, true) // 1s ICD per character

			c.a4Stacks++
			if c.a4Stacks > 4 {
				c.a4Stacks = 4
			}
			c.a4Expiry = c.Core.F + 8*60 // 8s duration, refreshed on new stack

			c.a4Apply()
			return false
		}, fmt.Sprintf("varka-a4-%v", ev))
	}
}

// a4Apply applies the A4 DMGP buff
func (c *char) a4Apply() {
	// Check if expired
	stacks := c.a4Stacks
	if c.Core.F >= c.a4Expiry {
		stacks = 0
		c.a4Stacks = 0
	}

	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = float64(stacks) * 0.075

	// C6: each stack also grants 20% CRIT DMG
	if c.Base.Cons >= 6 {
		m[attributes.CD] = float64(stacks) * 0.20
	}

	dur := c.a4Expiry - c.Core.F
	if dur <= 0 {
		return
	}

	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase(a4Key, dur),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			// Only applies to NA, CA, Azure Devour, FWA
			switch atk.Info.AttackTag {
			case attacks.AttackTagNormal, attacks.AttackTagExtra:
				return m, true
			case attacks.AttackTagElementalArt:
				// Only FWA, not Windbound Execution
				if atk.Info.Abil == "Four Winds' Ascension (Other)" ||
					atk.Info.Abil == "Four Winds' Ascension (Anemo)" {
					return m, true
				}
				return nil, false
			default:
				return nil, false
			}
		},
	})
}

// eleToStatP converts an element to its corresponding DMG% stat
func eleToStatP(ele attributes.Element) attributes.Stat {
	switch ele {
	case attributes.Pyro:
		return attributes.PyroP
	case attributes.Hydro:
		return attributes.HydroP
	case attributes.Electro:
		return attributes.ElectroP
	case attributes.Cryo:
		return attributes.CryoP
	case attributes.Anemo:
		return attributes.AnemoP
	default:
		return attributes.NoStat
	}
}
