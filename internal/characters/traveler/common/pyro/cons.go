package pyro

import (
	"github.com/genshinsim/gcsim/internal/template/nightsoul"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	c2Key = "traveller-c2"
)

func (c *Traveler) c1() {
	if c.Base.Cons < 1 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = 0.06
	n := make([]float64, attributes.EndStatType)
	n[attributes.DmgP] = 0.15
	for _, char := range c.Core.Player.Chars() {
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("traveller-c1", -1),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				if !c.StatusIsActive(nightsoul.NightsoulBlessingStatus) {
					return nil, false
				}
				if char.StatusIsActive(nightsoul.NightsoulBlessingStatus) {
					return n, true
				}
				return m, true
			},
		})
	}
}

func (c *Traveler) c2() {
	if c.Base.Cons < 2 {
		return
	}
	c2cb := func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)
		if ae.Info.ActorIndex != c.Core.Player.Active() {
			return false
		}
		if !c.StatusIsActive(c2Key) {
			return false
		}
		if c.c2Count >= 2 {
			return false
		}
		c.nightsoulState.GeneratePoints(14)
		c.c2Count++
		return false
	}
	c.Core.Events.Subscribe(event.OnOverload, c2cb, "traveller-c2")
	c.Core.Events.Subscribe(event.OnVaporize, c2cb, "traveller-c2")
	c.Core.Events.Subscribe(event.OnMelt, c2cb, "traveller-c2")
	c.Core.Events.Subscribe(event.OnBurning, c2cb, "traveller-c2")
	c.Core.Events.Subscribe(event.OnBurgeon, c2cb, "traveller-c2")
	c.Core.Events.Subscribe(event.OnSwirlPyro, c2cb, "traveller-c2")
	c.Core.Events.Subscribe(event.OnCrystallizePyro, c2cb, "traveller-c2")
}

func (c *Traveler) c4() {
	if c.Base.Cons < 4 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.PyroP] = 0.2
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag("traveler-c4", 9*60),
		AffectedStat: attributes.PyroP,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})
}
