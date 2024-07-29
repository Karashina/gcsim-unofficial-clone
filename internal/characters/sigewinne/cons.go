package sigewinne

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/core/player/shield"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const ()

func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}
	// add shield
	c.Core.Player.Shields.Add(&shield.Tmpl{
		ActorIndex: c.Index,
		Target:     c.Index,
		Src:        c.Core.F,
		ShieldType: shield.SigewinneC2,
		Name:       "Sigewinne C2",
		HP:         0.3 * c.MaxHP(),
		Ele:        attributes.Hydro,
		Expires:    c.Core.F + 15*60,
	})
}

func (c *char) c2Remove() {
	if c.Base.Cons < 2 {
		return
	}

	existingShield := c.Core.Player.Shields.Get(shield.TravelerHydroC4)
	if existingShield == nil {
		return
	}
	shd, _ := existingShield.(*shield.Tmpl)
	shd.Expires = c.Core.F + 1
}

func (c *char) c2Resist(a combat.AttackCB) {
	if c.Base.Cons < 2 {
		return
	}
	e, ok := a.Target.(*enemy.Enemy)
	if !ok {
		return
	}
	e.AddResistMod(combat.ResistMod{
		Base:  modifier.NewBaseWithHitlag("sigewinne-res-c2", 8*60),
		Ele:   attributes.Hydro,
		Value: -0.35,
	})
}

func (c *char) c6() {
	if c.Base.Cons < 6 {
		return
	}
	c.Core.Events.Subscribe(event.OnHeal, func(args ...interface{}) bool {
		hi := args[0].(*info.HealInfo)
		amount := args[2].(float64)
		if hi.Caller != c.Index {
			return false
		}
		if amount <= 0 {
			return false
		}

		c6bonuscr := min(0.2, c.MaxHP()/1000*0.004)
		c6bonuscd := min(1.1, c.MaxHP()/1000*0.022)

		m := make([]float64, attributes.EndStatType)
		m[attributes.CR] = c6bonuscr
		m[attributes.CD] = c6bonuscd
		c.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("sigewinne-c6-crit", 15*60),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				if atk.Info.AttackTag != attacks.AttackTagElementalBurst {
					return nil, false
				}
				return m, true
			},
		})

		return false
	}, "sigewinne-c6")
}
