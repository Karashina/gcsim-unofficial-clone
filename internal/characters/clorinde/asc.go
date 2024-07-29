package clorinde

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/gadget"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}

	a1cb := func(args ...interface{}) bool {
		c.a1stacks++
		if c.a1stacks >= 3 {
			c.a1stacks = 3
		}
		c.QueueCharTask(c.decreasea1, 15*60)

		c.a1buff = min(c.a1max, c.getTotalAtk()*c.a1mult*float64(c.a1stacks))

		c.Core.Log.NewEvent("Clorinde A1 Triggered", glog.LogCharacterEvent, c.Index).
			Write("a1 amt", c.a1buff)
		return false
	}

	a1cbNoGadget := func(args ...interface{}) bool {
		if _, ok := args[0].(*gadget.Gadget); ok {
			return false
		}

		return a1cb(args...)
	}

	c.Core.Events.Subscribe(event.OnOverload, a1cbNoGadget, "clorinde-a4")
	c.Core.Events.Subscribe(event.OnElectroCharged, a1cbNoGadget, "clorinde-a4")
	c.Core.Events.Subscribe(event.OnSuperconduct, a1cbNoGadget, "clorinde-a4")
	c.Core.Events.Subscribe(event.OnSwirlElectro, a1cbNoGadget, "clorinde-a4")
	c.Core.Events.Subscribe(event.OnCrystallizeElectro, a1cbNoGadget, "clorinde-a4")
	c.Core.Events.Subscribe(event.OnHyperbloom, a1cb, "clorinde-a4")
	c.Core.Events.Subscribe(event.OnQuicken, a1cbNoGadget, "clorinde-a4")
	c.Core.Events.Subscribe(event.OnAggravate, a1cbNoGadget, "clorinde-a4")
}

func (c *char) decreasea1() {
	if c.Base.Ascension < 1 {
		return
	}
	c.a1stacks--
	if c.a1stacks <= 0 {
		c.a1stacks = 0
	}
}

func (c *char) a4() {
	if c.Base.Ascension < 4 {
		c.healtobol = 0.8
		return
	}

	c.healtobol = 1

	c.a4stacks++
	if c.a4stacks >= 2 {
		c.a4stacks = 2
	}
	c.QueueCharTask(c.decreasea4, 15*60)

	c.a4buff = make([]float64, attributes.EndStatType)
	c.a4buff[attributes.CR] = 0.1 * float64(c.a4stacks)
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("clorinde-a4", 15*60),
		AffectedStat: attributes.CR,
		Amount: func() ([]float64, bool) {
			return c.a4buff, true
		},
	})
}

func (c *char) decreasea4() {
	if c.Base.Ascension < 1 {
		return
	}
	c.a1stacks--
	if c.a1stacks <= 0 {
		c.a1stacks = 0
	}
}
