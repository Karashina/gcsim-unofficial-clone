package skirk

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
)

const (
	a1IcdKey = "skirk-a1-icd"
)

func (c *char) a1(init bool) {
	if c.Base.Ascension < 1 {
		return
	}
	a1consume := func() {
		consumed := c.voidrift
		for range c.voidrift {
			c.c1()
			if c.c6count < 3 {
				c.c6count++
				c.AddStatus(c6key, 15*60, true)
			}
		}
		c.voidrift = 0
		if consumed > 0 {
			c.generateSerpentsSubtlety(8 * float64(consumed))
		}
	}

	a1generate := func(args ...interface{}) bool {
		if c.voidrift < 3 && !c.StatusIsActive(a1IcdKey) {
			c.AddStatus(a1IcdKey, 2.5*60, true)
			c.voidrift++
		}
		return false
	}

	if !init {
		a1consume()
	} else {
		c.Core.Events.Subscribe(event.OnFrozen, a1generate, "skirk-a1-frozen")
		c.Core.Events.Subscribe(event.OnSwirlCryo, a1generate, "skirk-a1-swirlcryo")
		c.Core.Events.Subscribe(event.OnSuperconduct, a1generate, "skirk-a1-superconduct")
		c.Core.Events.Subscribe(event.OnCrystallizeCryo, a1generate, "skirk-a1-crystallizecryo")

		c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
			if !c.onSevenPhaseFlash {
				return false
			}
			atk := args[1].(*combat.AttackEvent)
			if atk.Info.AttackTag != attacks.AttackTagExtra {
				return false
			}
			if atk.Info.ActorIndex != c.Index {
				return false
			}
			a1consume()
			return false
		}, "skirk-a1-chargehit")
	}
}

func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.Element != attributes.Cryo && atk.Info.Element != attributes.Hydro {
			return false
		}
		if atk.Info.ActorIndex == c.Index {
			return false
		}

		if !c.StatusIsActive(c.deathsCrossing[0]) {
			c.AddStatus(c.deathsCrossing[0], 20*60, true)
		} else if !c.StatusIsActive(c.deathsCrossing[1]) {
			c.AddStatus(c.deathsCrossing[1], 20*60, true)
		} else if !c.StatusIsActive(c.deathsCrossing[2]) {
			c.AddStatus(c.deathsCrossing[2], 20*60, true)
		}

		count := 0
		for i := 0; i < 3; i++ {
			if c.StatusIsActive(c.deathsCrossing[i]) {
				count++
			}
		}

		switch count {
		case 1:
			c.a4BuffNA = 1.10
			c.a4BuffQ = 1.105
		case 2:
			c.a4BuffNA = 1.20
			c.a4BuffQ = 1.15
		case 3:
			c.a4BuffNA = 1.70
			c.a4BuffQ = 1.60
		default:
			c.a4BuffNA = 1
			c.a4BuffQ = 1
		}

		return false
	}, "skirk-a4")
}
