package kachina

import (
	"math"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/player/shield"
	"github.com/genshinsim/gcsim/pkg/reactable"
)

const (
	c1IcdKey = "kachina-c1-icd"
	c6IcdKey = "kachina-c6-icd"
)

func (c *char) c1shard() {
	if c.Base.Cons < 1 {
		return
	}
	Area := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 8)
	// suck in all crystallize shard
	for _, g := range c.Core.Combat.Gadgets() {
		cs, ok := g.(*reactable.CrystallizeShard)
		// skip if no shard
		if !ok {
			continue
		}
		// skip if shard not in area
		if !cs.IsWithinArea(Area) {
			continue
		}
		// approximate sucking in as 0.4m per frame
		distance := cs.Pos().Distance(Area.Shape.Pos())
		travel := int(math.Ceil(distance / 0.4))
		// special check to account for edge case if shard just spawned and will arrive before it can be picked up
		if c.Core.F+travel < cs.EarliestPickup {
			continue
		}
		c.Core.Tasks.Add(func() {
			cs.AddShieldKillShard()
		}, travel)
	}
}

func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}
	c.Core.Events.Subscribe(event.OnShielded, func(args ...interface{}) bool {
		// Check shield
		shd := args[0].(shield.Shield)
		if shd.Type() != shield.Crystallize {
			return false
		}
		if c.StatusIsActive(c1IcdKey) {
			return false
		}
		c.AddEnergy("kachina-c1-energy", 3)
		c.AddStatus(c1IcdKey, 5*60, true)
		c.Core.Log.NewEvent("Energy gained from Crystallise", glog.LogCharacterEvent, c.Index)
		return false
	}, "shrapnel-gain")
}

func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}
	if !c.StatusIsActive(skillKey) {
		c.AddStatus(skillKey, -1, true)
		c.OnNightsoul = true
		c.AddNightsoul("kachina-c2", 20)
		c.newTwirly()
	} else {
		c.AddNightsoul("kachina-c2", 20)
	}
}

func (c *char) c6() {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "This Time, I've Gotta Win (C6)",
		AttackTag:  attacks.AttackTagNone,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		Element:    attributes.Geo,
		Durability: 25,
		Mult:       2,
		UseDef:     true,
	}

	c.Core.Events.Subscribe(event.OnShielded, func(args ...interface{}) bool {
		if c.StatusIsActive(c6IcdKey) {
			return false
		}
		c.AddStatus(c6IcdKey, 5*60, true)
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player().Pos(), nil, 3.5), 0, 1)
		return false
	}, "kachina-shield-gain")

	c.Core.Events.Subscribe(event.OnShieldBreak, func(args ...interface{}) bool {
		if c.StatusIsActive(c6IcdKey) {
			return false
		}
		c.AddStatus(c6IcdKey, 5*60, true)
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player().Pos(), nil, 3.5), 0, 1)
		return false
	}, "kachina-shield-broken")

}
