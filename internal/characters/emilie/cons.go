package emilie

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	c1scentsICDKey = "emilie-c1-scents-icd"
	c6Key          = "emilie-c6"
	c6ICDKey       = "emilie-c6-icd"
)

func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = 0.2
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("emilie-c1", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag == attacks.AttackTagNone || atk.Info.AttackTag == attacks.AttackTagElementalArt {
				return m, true
			}
			return nil, false
		},
	})
}

func (c *char) c1dendro() {
	if c.Base.Cons < 1 {
		return
	}
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		t, ok := args[0].(*enemy.Enemy)
		if !ok {
			return false
		}
		if t.IsBurning() && atk.Info.Element == attributes.Dendro {
			c.c1handle()
		}
		return false
	}, "emilie-c1-dendro")
}

func (c *char) c1burn() {
	if c.Base.Cons < 1 {
		return
	}
	c.Core.Events.Subscribe(event.OnBurning, func(args ...interface{}) bool {
		c.c1handle()
		return false
	}, "emilie-c1-burn")
}

func (c *char) c1handle() {
	if c.Base.Cons < 1 {
		return
	}
	if !c.LCActive {
		c.Scents = 0
		c.LCLevel = 0
	}
	if c.Scents < 2 {
		c.Scents++
		c.AddStatus(c1scentsICDKey, 2.9*60, false)
	}
	if c.Scents >= 2 && c.LCLevel != 1 {
		c.LCLevel = 1
	}
	if c.Scents >= 2 {
		c.a1()
		c.Scents = 0
	}
}

func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		dmg := args[2].(float64)
		t, ok := args[0].(*enemy.Enemy)
		if !ok {
			return false
		}
		if dmg == 0 {
			return false
		}

		if atk.Info.ActorIndex != c.Index {
			return false
		}

		if atk.Info.AttackTag == attacks.AttackTagNone || atk.Info.AttackTag == attacks.AttackTagElementalBurst || atk.Info.AttackTag == attacks.AttackTagElementalArt {
			t.AddResistMod(combat.ResistMod{
				Base:  modifier.NewBaseWithHitlag("emilie-c2-dendro", 10*60),
				Ele:   attributes.Dendro,
				Value: -0.30,
			})
		}

		return false
	}, "emilie-c2")
}

func (c *char) c6() {
	if c.Base.Cons < 6 {
		return
	}
	if !c.StatusIsActive(c6Key) {
		return
	}
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)
		if ae.Info.AttackTag == attacks.AttackTagNormal || ae.Info.AttackTag == attacks.AttackTagExtra {
			ae.Info.Element = attributes.Dendro
			ae.Info.FlatDmg += c.TotalAtk() * 3
			c.c6Count += 1
			if c.c6Count == 4 {
				c.DeleteStatus(c6Key)
				c.c6Count = 0
			}
		}
		return false
	}, "emilie-c6")
}

func (c *char) c6cb(a combat.AttackCB) {
	if c.Base.Cons < 6 {
		return
	}
	if c.StatusIsActive(c6ICDKey) {
		return
	}
	c.AddStatus(c6ICDKey, 12*60, true)
	c.AddStatus(c6Key, 5*60, true)
}
