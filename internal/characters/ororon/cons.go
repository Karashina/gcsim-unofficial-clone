package ororon

import (
	"sort"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/core/targets"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	c1Key  = "ororon-c1"
	c2Key  = "ororon-c2"
	c6Key1 = "ororon-c6-stack-1"
	c6Key2 = "ororon-c6-stack-2"
	c6Key3 = "ororon-c6-stack-3"
)

func (c *char) c1cb(a combat.AttackCB) {
	if c.Base.Cons < 1 {
		return
	}
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	c.AddStatus(c1Key, 12*60, true)
}

func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}
	c1buff := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBaseWithHitlag("Trails Amidst the Forest Fog (C1)", -1),
		Amount: func(atk *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
			if !c.StatusIsActive(c1Key) {
				return nil, false
			}
			if atk.Info.Abil != "Hypersense (A1)" {
				return nil, false
			}
			c1buff[attributes.DmgP] = 0.5
			return c1buff, true
		},
	})
}

func (c *char) c2Init() {
	if c.Base.Cons < 2 {
		return
	}
	c.c2stacks = 0
	c.c2buff[attributes.ElectroP] = 0.08
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag(c2Key, 9*60),
		AffectedStat: attributes.ElectroP,
		Extra:        true,
		Amount: func() ([]float64, bool) {
			return c.c2buff, true
		},
	})
}

func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if !c.StatModIsActive(c2Key) {
			return false
		}
		if atk.Info.ActorIndex != c.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagElementalBurst {
			return false
		}
		c.c2stacks++
		c.c2buff[attributes.ElectroP] = min(0.32, 0.08+0.08*float64(c.c2stacks))
		remain := c.StatusDuration(c2Key)
		c.DeleteStatMod(c2Key)
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(c2Key, remain),
			AffectedStat: attributes.ElectroP,
			Extra:        true,
			Amount: func() ([]float64, bool) {
				return c.c2buff, true
			},
		})
		return false
	}, "ororon-c2")
}

func (c *char) c6cb(a combat.AttackCB) {
	if c.Base.Cons < 6 {
		return
	}
	c.c6buff[attributes.ATKP] = 0.1

	statusDurations := []struct {
		key      string
		duration int
	}{
		{c6Key1, c.StatusDuration(c6Key1)},
		{c6Key2, c.StatusDuration(c6Key2)},
		{c6Key3, c.StatusDuration(c6Key3)},
	}

	sort.Slice(statusDurations, func(i, j int) bool {
		return statusDurations[i].duration < statusDurations[j].duration
	})

	oldestStatus := statusDurations[0]
	if !c.StatusIsActive(oldestStatus.key) {
		c.AddStatus(oldestStatus.key, 9*60, true)
		for _, chr := range c.Core.Player.Chars() {
			chr.AddAttackMod(character.AttackMod{
				Base: modifier.NewBaseWithHitlag(oldestStatus.key, 9*60),
				Amount: func(atk *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
					if atk.Info.ActorIndex != c.Core.Player.Active() {
						return nil, false
					}
					return c.c6buff, true
				},
			})
		}
	}
}

func (c *char) c6() {
	if c.Base.Cons < 6 {
		return
	}
	if c.Base.Ascension < 1 {
		return
	}

	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Hypersense (C6)",
		AttackTag:      attacks.AttackTagNone,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagElementalArt,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypePierce,
		Element:        attributes.Electro,
		Durability:     25,
		Mult:           1.6 * 2,
	}

	area := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 10)
	enemies := c.Core.Combat.RandomEnemiesWithinArea(area, nil, 4)

	for _, e := range enemies {
		c.Core.QueueAttack(
			ai,
			combat.NewSingleTargetHit(e.Key()),
			a1hitmark,
			a1hitmark,
			c.c6cb,
		)
	}
}
