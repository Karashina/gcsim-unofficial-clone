package escoffier

import (
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	c2Key     = "escoffier-c2"
	c6IcdKey  = "escoffier-c6-icd"
	c6Hitmark = 10
)

func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}
	c.Core.Events.Subscribe(event.OnActionExec, func(args ...interface{}) bool {
		idx := args[0].(int)
		act := args[1].(action.Action)
		if idx != c.Index {
			return false
		}
		if act != action.ActionSkill && act != action.ActionBurst {
			return false
		}

		m := make([]float64, attributes.EndStatType)
		m[attributes.CD] = 0.60
		for _, char := range c.Core.Player.Chars() {
			char.AddAttackMod(character.AttackMod{
				Base: modifier.NewBaseWithHitlag("escoffier-c1-cryocd", 15*60),
				Amount: func(ae *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
					if ae.Info.Element != attributes.Cryo {
						return nil, false
					}
					return m, true
				},
			})
		}
		return false
	}, "escoffier-c1")
}

func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		if !c.StatusIsActive(c2Key) {
			return false
		}
		if c.c2count >= 5 {
			return false
		}
		ae := args[1].(*combat.AttackEvent)
		if ae.Info.ActorIndex == c.Index {
			return false
		}
		switch ae.Info.AttackTag {
		case attacks.AttackTagNormal:
		case attacks.AttackTagExtra:
		case attacks.AttackTagPlunge:
		case attacks.AttackTagElementalArt:
		case attacks.AttackTagElementalBurst:
		default:
			return false
		}

		dmgAdded := 2.4 * c.TotalAtk()
		ae.Info.FlatDmg += dmgAdded
		c.c2count++

		c.Core.Log.NewEvent("escoffier c2 adding damage", glog.LogPreDamageMod, ae.Info.ActorIndex).
			Write("damage_added", dmgAdded)

		return false
	}, "escoffier-c2")
}

func (c *char) c4() bool {
	if c.Base.Cons < 4 {
		return false
	}
	if c.c4count >= 7 {
		return false
	}
	if c.Core.Rand.Float64() < c.Stat(attributes.CR) {
		c.c4count++
		c.AddEnergy("escoffier-c4", 2)
		return true
	}
	return false
}

func (c *char) c6() {
	if c.Base.Cons < 6 {
		return
	}
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		if c.StatusIsActive(c6IcdKey) {
			return false
		}
		if c.c6count >= 6 {
			return false
		}
		ae := args[1].(*combat.AttackEvent)
		if ae.Info.ActorIndex != c.Core.Player.Active() {
			return false
		}
		if ae.Info.AttackTag != attacks.AttackTagNormal && ae.Info.AttackTag != attacks.AttackTagExtra && ae.Info.AttackTag != attacks.AttackTagPlunge {
			return false
		}

		c.AddStatus(c6IcdKey, 0.5*60, false)
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Tea Parties Bursting With Color (C6)",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagElementalArt,
			ICDGroup:   attacks.ICDGroupEscoffierCookingMek,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Cryo,
			Durability: 25,
			Mult:       5,
		}
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), geometry.Point{Y: 2.6}, 4.5),
			c6Hitmark,
			c6Hitmark,
		)
		c.c6count++

		return false
	}, "escoffier-c6")
}
