package chasca

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/core/targets"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	c2ICDKey = "chasca-c2-icd"
	c4ICDKey = "chasca-c4-icd"
	c6Key    = "chasca-c6"
	c6ICDKey = "chasca-c6-icd"
)

func (c *char) c2CB(a combat.AttackCB) {
	if c.Base.Cons < 2 {
		return
	}
	if a.AttackEvent.Info.Element == attributes.Anemo {
		return
	}
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(c2ICDKey) {
		return
	}

	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Muzzle, the Searing Smoke (C2)",
		AttackTag:      attacks.AttackTagExtra,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagNone,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeDefault,
		Mult:           4,
		Element:        a.AttackEvent.Info.Element,
		Durability:     25,
	}

	c.Core.QueueAttack(
		ai,
		combat.NewSingleTargetHit(c.Core.Combat.PrimaryTarget().Key()),
		0,
		0,
	)

	c.AddStatus(c2ICDKey, particleICD, true)
}

func (c *char) c4CB(a combat.AttackCB) {
	if c.Base.Cons < 4 {
		return
	}
	c.AddEnergy("chasca-c4", 1.5)

	if a.AttackEvent.Info.Abil != "Radiant Soulseeker Shell DMG (Q)" {
		return
	}
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(c4ICDKey) {
		return
	}

	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Sparks, the Sudden Shot (C4)",
		AttackTag:      attacks.AttackTagExtra,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagNone,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeDefault,
		Mult:           4,
		Element:        a.AttackEvent.Info.Element,
		Durability:     25,
	}

	c.Core.QueueAttack(
		ai,
		combat.NewSingleTargetHit(c.Core.Combat.PrimaryTarget().Key()),
		0,
		0,
	)

	c.AddStatus(c4ICDKey, particleICD, true)
}

func (c *char) c6CDbuff() {
	if c.Base.Cons < 6 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBaseWithHitlag("Showdown, the Glory of Battle (C6)", -1),
		Amount: func(atk *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
			if !c.StatusIsActive(c6Key) {
				return nil, false
			}
			if atk.Info.Abil != "Shining Shadowhunt Shell DMG (E)" && atk.Info.Abil != "Shadowhunt Shell DMG (E)" {
				return nil, false
			}
			m[attributes.CD] = 1.2
			return m, true
		},
	})
}

func (c *char) removec6() {
	if c.Base.Cons < 6 {
		return
	}
	c.DeleteStatus(c6Key)
}
