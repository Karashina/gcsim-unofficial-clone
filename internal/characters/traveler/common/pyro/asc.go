package pyro

import (
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
)

const a4IcdKey = "traveller-a4-icd"

func (c *Traveler) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	a4cb := func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)
		if ae.Info.ActorIndex != c.Core.Player.Active() {
			return false
		}
		if c.StatusIsActive(a4IcdKey) {
			return false
		}
		c.AddStatus(a4IcdKey, 12*60, true)
		c.AddEnergy("traveller-a4-reaction", 5)
		return false
	}
	c.Core.Events.Subscribe(event.OnNightsoulBurst, func(args ...interface{}) bool {
		c.AddEnergy("traveller-a4-nightsoul", 4)
		return false
	}, "traveller-a4-nightsoul-check")
	c.Core.Events.Subscribe(event.OnOverload, a4cb, "traveller-a4")
	c.Core.Events.Subscribe(event.OnVaporize, a4cb, "traveller-a4")
	c.Core.Events.Subscribe(event.OnMelt, a4cb, "traveller-a4")
	c.Core.Events.Subscribe(event.OnBurning, a4cb, "traveller-a4")
	c.Core.Events.Subscribe(event.OnBurgeon, a4cb, "traveller-a4")
	c.Core.Events.Subscribe(event.OnSwirlPyro, a4cb, "traveller-a4")
	c.Core.Events.Subscribe(event.OnCrystallizePyro, a4cb, "traveller-a4")
}
