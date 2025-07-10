package skirk

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	c1Hitmark = 1
	c6key     = "skirk-c6"
)

func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Far to Fall (C1)",
		AttackTag:  attacks.AttackTagExtra,
		ICDTag:     attacks.ICDTagExtraAttack,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeSlash,
		Element:    attributes.Cryo,
		Durability: 25,
		Mult:       5,
	}
	ap := combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: -1}, 2, 2)
	c.Core.QueueAttack(ai, ap, c1Hitmark, c1Hitmark, c.particleCB)
}

func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.ATKP] = 0.7
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag("skirk-c2", 12.5*60),
		AffectedStat: attributes.ATKP,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})
}

func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag("skirk-c4", -1),
		AffectedStat: attributes.ATKP,
		Amount: func() ([]float64, bool) {
			count := 0
			for i := 0; i < 3; i++ {
				if c.StatusIsActive(c.deathsCrossing[i]) {
					count++
				}
			}
			switch count {
			case 1:
				m[attributes.ATKP] = 0.1
			case 2:
				m[attributes.ATKP] = 0.2
			case 3:
				m[attributes.ATKP] = 0.4
			default:
				m[attributes.ATKP] = 0
			}
			return m, true
		},
	})

}

func (c *char) c6(typ string) {
	if c.Base.Cons < 6 {
		return
	}
	if !c.StatusIsActive(c6key) {
		c.c6count = 0
		return
	}
	if c.c6count <= 0 {
		return
	}
	if typ == "burst" {
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "To the Source (C6 - Burst)",
			AttackTag:  attacks.AttackTagElementalBurst,
			ICDTag:     attacks.ICDTagElementalBurst,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeSlash,
			Element:    attributes.Cryo,
			Durability: 25,
			Mult:       7.5,
		}
		ap := combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: -1}, 2, 2)
		c.c6count = 0
		for range c.c6count {
			c.Core.QueueAttack(ai, ap, c1Hitmark, c1Hitmark, c.particleCB)
		}
	}
	if typ == "na" {
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "To the Source (C6 - NA)",
			AttackTag:  attacks.AttackTagNormal,
			ICDTag:     attacks.ICDTagNormalAttack,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeSlash,
			Element:    attributes.Cryo,
			Durability: 25,
			Mult:       1.8,
		}
		ap := combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: -1}, 2, 2)
		c.c6count--
		for i := 0; i < 3; i++ {
			c.Core.QueueAttack(ai, ap, c1Hitmark, c1Hitmark, c.particleCB)
		}
	}
}
