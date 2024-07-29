package clorinde

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	c1ICDkey = "clorinde-c1-icd"
	c6ICDkey = "clorinde-c6-icd"
)

func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}
	if c.StatusIsActive(c1ICDkey) {
		return
	}
	if !c.StatusIsActive(skillBuffKey) {
		return
	}
	c.AddStatus(c1ICDkey, 1.2*60, true)
	c1ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Clorinde C1",
		AttackTag:          attacks.AttackTagNone,
		ICDTag:             attacks.ICDTagClorindeCons,
		ICDGroup:           attacks.ICDGroupClorindeCons,
		StrikeType:         attacks.StrikeTypePierce,
		Element:            attributes.Electro,
		Durability:         25,
		Mult:               0.3,
		FlatDmg:            c.a1buff,
		HitlagFactor:       0,
		CanBeDefenseHalted: true,
	}

	c.Core.QueueAttack(
		c1ai,
		combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: -0.3}, 1.2, 4.5),
		15,
		15,
		c.particleCB,
	)
	c.Core.QueueAttack(
		c1ai,
		combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: -0.3}, 1.2, 4.5),
		21,
		21,
		c.particleCB,
	)
}

func (c *char) c4() {
	c.c4mult = c.CurrentHPDebt()*2/c.MaxHP() + 1
	c.c4max = 2

	m := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("clorinde-c4", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagElementalBurst {
				return nil, false
			}
			m[attributes.DmgP] = min(c.c4max, c.c4mult)
			return m, true
		},
	})
}

func (c *char) c6init() {
	c.c6count = 0
	m := make([]float64, attributes.EndStatType)
	n := make([]float64, attributes.EndStatType)
	m[attributes.CR] = 0.1
	n[attributes.CD] = 0.7
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("clorinde-c6-buff-cr", 12*60),
		AffectedStat: attributes.CR,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("clorinde-c6-buff-cd", 12*60),
		AffectedStat: attributes.CD,
		Amount: func() ([]float64, bool) {
			return n, true
		},
	})
}

func (c *char) c6() {
	if c.StatusIsActive(c6ICDkey) {
		return
	}
	if c.c6count > 6 {
		return
	}
	c.AddStatus(c6ICDkey, 60, true)
	c.c6count++
	c6ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Clorinde C6",
		AttackTag:          attacks.AttackTagNone,
		ICDTag:             attacks.ICDTagClorindeCons,
		ICDGroup:           attacks.ICDGroupClorindeCons,
		StrikeType:         attacks.StrikeTypePierce,
		Element:            attributes.Electro,
		Durability:         25,
		Mult:               2,
		FlatDmg:            c.a1buff,
		HitlagFactor:       0,
		CanBeDefenseHalted: true,
	}
	c.c1()
	c.Core.QueueAttack(
		c6ai,
		combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: -0.3}, 1.2, 4.5),
		15,
		15,
		c.particleCB,
	)
}
