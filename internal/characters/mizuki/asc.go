package mizuki

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const a1IcdKey = "mizuki-a1-icd"

func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	a1cb := func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)

		if !c.StatusIsActive(skillKey) {
			return false
		}

		if c.Tags[skillKey] <= 0 {
			return false
		}

		if ae.Info.ActorIndex != c.Index {
			return false
		}

		if c.StatusIsActive(a1IcdKey) {
			return false
		}

		c.AddStatus(a1IcdKey, 0.3*60, true)
		c.Tags[skillKey] -= 1
		c.ExtendStatus(skillKey, 2.5*60)

		return false
	}

	c.Core.Events.Subscribe(event.OnSwirlCryo, a1cb, "mizuki-a1")
	c.Core.Events.Subscribe(event.OnSwirlElectro, a1cb, "mizuki-a1")
	c.Core.Events.Subscribe(event.OnSwirlHydro, a1cb, "mizuki-a1")
	c.Core.Events.Subscribe(event.OnSwirlPyro, a1cb, "mizuki-a1")
}

func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)

		if !c.StatusIsActive(skillKey) {
			return false
		}

		//TODO: is it possible to trigger a4 with her swirl?
		if ae.Info.ActorIndex == c.Index {
			return false
		}

		switch ae.Info.Element {
		case attributes.Electro:
		case attributes.Hydro:
		case attributes.Pyro:
		case attributes.Cryo:
		default:
			return false
		}

		m := make([]float64, attributes.EndStatType)
		m[attributes.EM] = 100
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("mizuki-a4", 4*60),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})

		return false
	}, "mizuki-a4")
}
