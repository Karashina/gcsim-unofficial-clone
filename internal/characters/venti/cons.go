package venti

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// C1 (original): Fires 2 additional arrows per Aimed Shot, each dealing 33% of the original arrow's DMG.
// C1 (hexerei addition, witch's challenge required):
//
//	Stormwind Arrows also fire 2 split tracking arrows, each dealing 20% of the original arrow's DMG.
//	This effect can trigger once per 0.25s (15 frames).
func (c *char) c1(ai combat.AttackInfo, hitmark, travel int) {
	ai.Abil += " (C1)"
	ai.Mult /= 3.0
	for i := 0; i < 2; i++ {
		c.Core.QueueAttack(
			ai,
			combat.NewBoxHit(
				c.Core.Combat.Player(),
				c.Core.Combat.PrimaryTarget(),
				geometry.Point{Y: -0.5},
				0.1,
				1,
			),
			hitmark,
			hitmark+travel,
		)
	}
}

// makeC1StormwindSplitCB returns an AttackCBFunc that fires 2 extra tracking arrows
// at 20% of the Stormwind Arrow's DMG on hit.
// Requires: C1, hexerei mode, burst eye active. ICD: 0.25s (15 frames).
func (c *char) makeC1StormwindSplitCB() combat.AttackCBFunc {
	return func(a combat.AttackCB) {
		if c.Base.Cons < 1 {
			return
		}
		if !c.isHexerei || !c.hasHexBonus {
			return
		}
		if c.Core.F-c.lastStormwindSplit < 15 { // 0.25s ICD
			return
		}
		c.lastStormwindSplit = c.Core.F
		// Derive split mult from the actual hit's AttackEvent
		splitMult := a.AttackEvent.Info.Mult * 0.20
		splitAI := a.AttackEvent.Info
		splitAI.Abil += " (C1 Split)"
		splitAI.Mult = splitMult
		for i := 0; i < 2; i++ {
			c.Core.QueueAttack(
				splitAI,
				combat.NewBoxHit(
					c.Core.Combat.Player(),
					c.Core.Combat.PrimaryTarget(),
					geometry.Point{Y: -0.5},
					0.1,
					1,
				),
				0,
				1,
			)
		}
	}
}

// C2 (original): Skyward Sonnet decreases opponents' Anemo RES and Physical RES by 12% for 10s.
// C2 (hexerei addition): Press Skyward Sonnet deals 300% of original damage (hexerei only).
//
//	The 300% multiplier is applied in skill.go when hexerei is active.
func (c *char) c2(a combat.AttackCB) {
	if c.Base.Cons < 2 {
		return
	}
	e, ok := a.Target.(*enemy.Enemy)
	if !ok {
		return
	}
	e.AddResistMod(combat.ResistMod{
		Base:  modifier.NewBaseWithHitlag("venti-c2-anemo", 600),
		Ele:   attributes.Anemo,
		Value: -0.12,
	})
	e.AddResistMod(combat.ResistMod{
		Base:  modifier.NewBaseWithHitlag("venti-c2-phys", 600),
		Ele:   attributes.Physical,
		Value: -0.12,
	})
}

// C4 (original): When Venti picks up an Elemental Orb or Particle, he receives Anemo DMG +25% for 10s.
// C4 (hexerei addition): After Venti uses Skyward Sonnet or Wind's Grand Ode, Venti and other
//
//	active party members gain Anemo DMG +25% for 10s (hexerei only, initialized in venti.go Init).
func (c *char) c4Old() {
	c4bonus := make([]float64, attributes.EndStatType)
	c4bonus[attributes.AnemoP] = 0.25
	c.Core.Events.Subscribe(event.OnParticleReceived, func(args ...interface{}) bool {
		if c.Core.Player.Active() != c.Index {
			return false
		}
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("venti-c4-old", 600),
			AffectedStat: attributes.AnemoP,
			Amount: func() ([]float64, bool) {
				return c4bonus, true
			},
		})
		return false
	}, "venti-c4-old")
}

func (c *char) c4New() {
	if !c.isHexerei {
		return
	}
	for _, ch := range c.Core.Player.Chars() {
		m := make([]float64, attributes.EndStatType)
		m[attributes.AnemoP] = 0.25
		ch.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("venti-c4-hex", 10*60),
			AffectedStat: attributes.NoStat,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}
}

// C6: (Unlocked by completing Witch's Challenge)
// Targets hit by Wind's Grand Ode have their Anemo RES decreased by 20%.
// If Elemental Absorption occurred, that element's RES is also decreased by 20%.
// Additionally, Venti's CRIT DMG against these enemies is increased by 100%.
func (c *char) c6(ele attributes.Element) func(a combat.AttackCB) {
	return func(a combat.AttackCB) {
		e, ok := a.Target.(*enemy.Enemy)
		if !ok {
			return
		}
		e.AddResistMod(combat.ResistMod{
			Base:  modifier.NewBaseWithHitlag("venti-c6-"+ele.String(), 600),
			Ele:   ele,
			Value: -0.20,
		})
	}
}

// c6AttackModInit adds a persistent AttackMod giving Venti +100% CRIT DMG against enemies
// that have the Anemo RES debuff applied by c6 (hexerei addition only).
func (c *char) c6AttackModInit() {
	if !c.isHexerei {
		return
	}
	m := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("venti-c6-cd", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.ActorIndex != c.Index {
				return nil, false
			}
			e, ok := t.(*enemy.Enemy)
			if !ok {
				return nil, false
			}
			// Check if target was hit by burst (has anemo res debuff)
			if !e.StatusIsActive("venti-c6-anemo") {
				return nil, false
			}
			for i := range m {
				m[i] = 0
			}
			m[attributes.CD] = 1.0
			return m, true
		},
	})
}

// hexAttackEnabled returns true when the hexerei normal attack passive should be active.
// Requires: hexerei mode, 2+ hexerei party members, and burst eye active.
// This effect is unlocked by completing the Witch's Challenge (hexerei flag), not constellation-gated.
func (c *char) hexAttackEnabled() bool {
	return c.isHexerei && c.hasHexBonus && c.Core.F < c.burstEnd
}

// makeHexNormalCB returns an AttackCBFunc that extends the burst eye duration and
// reduces burst CD on normal attack hits (hexerei passive).
func (c *char) makeHexNormalCB() combat.AttackCBFunc {
	return func(a combat.AttackCB) {
		if c.normalHexCount >= 2 {
			return
		}
		if c.Core.F-c.lastHexTrigger < 6 { // 0.1s ICD = 6 frames
			return
		}
		c.lastHexTrigger = c.Core.F
		c.normalHexCount++
		// Extend burst eye duration by 1s (60 frames)
		c.burstEnd += 60
		// Reduce burst CD by 0.5s (30 frames)
		c.ReduceActionCooldown(action.ActionBurst, 30)
	}
}
