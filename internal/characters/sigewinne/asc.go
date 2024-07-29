package sigewinne

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	a1BuffKey = "sigewinne-a1"
)

func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	c.SetTag(a1BuffKey, 10)
	c.AddStatus(a1BuffKey, 18*60, true)
	a1hydrobuff := make([]float64, attributes.EndStatType)
	a1hydrobuff[attributes.HydroP] = 0.08
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("sigewinne-a1-hydroP", 18*60),
		AffectedStat: attributes.HydroP,
		Amount: func() ([]float64, bool) {
			return a1hydrobuff, true
		},
	})
}

func (c *char) a1Proc() {
	if c.Base.Ascension < 1 {
		return
	}
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)

		if ae.Info.ActorIndex == c.Index {
			return false
		}
		if ae.Info.ActorIndex == c.Core.Player.ActiveChar().Index {
			return false
		}
		if ae.Info.AttackTag != attacks.AttackTagElementalArt && ae.Info.AttackTag != attacks.AttackTagElementalArtHold {
			return false
		}
		// do nothing if buff gone or count gone
		if c.Tags[a1BuffKey] <= 0 {
			return false
		}
		if !c.StatusIsActive(a1BuffKey) {
			return false
		}

		var a1Buff float64
		if c.Base.Cons < 2 {
			a1Buff = min(2800, (c.MaxHP()-30000)/1000*80)
		} else {
			a1Buff = min(3500, (c.MaxHP()-30000)/1000*100)
		}

		ae.Info.FlatDmg += a1Buff

		c.Tags[a1BuffKey] -= 1

		c.Core.Log.NewEvent("sigewinne a1 adding damage", glog.LogPreDamageMod, ae.Info.ActorIndex).
			Write("damage_added", a1Buff).
			Write("stacks_remaining", c.Tags[a1BuffKey]).
			Write("buff amt", a1Buff)

		return false
	}, "sigewinne-a1-buff")
}

func (c *char) a4() {
	if c.Base.Ascension < 4 {
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

		var totaldebt float64

		for _, char := range c.Core.Player.Chars() {
			totaldebt += char.CurrentHPDebt()
		}

		healbonus := min(0.3, totaldebt/1000*0.03)

		c.AddHealBonusMod(character.HealBonusMod{
			Base: modifier.NewBase("sigewinne-a4-heal-buff", -1),
			Amount: func() (float64, bool) {
				return healbonus, false
			},
		})

		return false
	}, "sigewinne-a4")
}
