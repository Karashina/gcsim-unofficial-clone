package kinich

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/enemy"
)

const (
	A1MarkKey = "kinich-desolation-mark"
	A1ICDKey  = "kinich-desolation-icd"
	A4DurKey  = "kinich-a4-duration"
)

func (c *char) a1mark() {
	if c.Base.Ascension < 1 {
		return
	}
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		if !c.StatusIsActive(skillKey) {
			return false
		}
		t, ok := args[0].(*enemy.Enemy)
		atk := args[1].(*combat.AttackEvent)
		if !ok {
			return false
		}
		if atk.Info.ActorIndex != c.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagElementalArt {
			return false
		}

		t.AddStatus(A1MarkKey, -1, true)

		return false
	}, "kinich-desolation-mark")
}

func (c *char) a1regen() {
	if c.Base.Ascension < 1 {
		return
	}
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		if !c.StatusIsActive(skillKey) {
			return false
		}
		if c.StatusIsActive(A1ICDKey) {
			return false
		}
		t, ok := args[0].(*enemy.Enemy)
		atk := args[1].(*combat.AttackEvent)
		if !ok {
			return false
		}
		if !t.StatusIsActive(A1MarkKey) {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagBurningDamage && atk.Info.AttackTag != attacks.AttackTagBurgeon {
			return false
		}

		c.AddNightsoul("kinich-a1", 7)
		c.AddStatus(A1ICDKey, 0.8*60, true)

		return false
	}, "kinich-a1-nightsoul-regen")
}

func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	if !c.StatusIsActive(A4DurKey) {
		c.a4stacks = 0
		c.a4buff = 0
	}
	c.Core.Events.Subscribe(event.OnNightsoulBurst, func(args ...interface{}) bool {
		if c.a4stacks < 2 {
			c.a4stacks++
		}
		c.a4buff = 3.2 * float64(c.a4stacks)
		c.AddStatus(A4DurKey, 15*60, true)

		return false
	}, "kinich-a4")
}
