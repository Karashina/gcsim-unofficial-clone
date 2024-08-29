package mualani

import (
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/targets"
)

func (c *char) a1cb(a combat.AttackCB) {
	if c.Base.Ascension < 1 {
		return
	}
	if a.AttackEvent.Info.Abil != "Sharky's Surging Bite DMG" {
		return
	}
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if !c.StatusIsActive(skillKey) {
		return
	}
	if c.pufferCount >= 2 {
		return
	}
	c.QueueCharTask(c.a1, 117)
	c.pufferCount++
}

func (c *char) a1() {
	c.c4energy()
	if !c.StatusIsActive(skillKey) {
		return
	}
	c.AddNightsoul("mualani-a1", 20)
	c.c2Puffer()
}

func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	c.Core.Events.Subscribe(event.OnNightsoulBurst, func(args ...interface{}) bool {
		if c.a4stacks < 3 {
			c.a4stacks++
		}
		c.a4buff = c.MaxHP() * 0.15 * float64(c.a4stacks)
		return false
	}, "mualani-a4")
}
