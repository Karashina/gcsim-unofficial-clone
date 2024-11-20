package chasca

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}

	c.a1Prob = 0
	a4buff := 0.0
	c1mod := 0.0

	if c.Base.Cons >= 1 {
		c1mod = 0.333
	}

	switch c.typeCount {
	case 0:
		c.a1Prob = 0 + c1mod
		if c.Base.Cons >= 2 {
			a4buff = 0.15
		}
	case 1:
		c.a1Prob = 0.33 + c1mod
		a4buff = 0.15
		if c.Base.Cons >= 2 {
			a4buff = 0.35
		}
	case 2:
		c.a1Prob = 0.667 + c1mod
		a4buff = 0.35
		if c.Base.Cons >= 2 {
			a4buff = 0.65
		}
	case 3:
		c.a1Prob = 1
		a4buff = 0.65
	default:
		c.a1Prob = 0
	}

	m := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBaseWithHitlag("Bullet Trick (A1)", -1),
		Amount: func(atk *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
			if atk.Info.Abil != "Shining Shadowhunt Shell DMG (E)" {
				return nil, false
			}
			m[attributes.DmgP] = a4buff
			return m, true
		},
	})
}

func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	c.Core.Events.Subscribe(event.OnNightsoulBurst, func(args ...interface{}) bool {
		ai := combat.AttackInfo{
			ActorIndex:     c.Index,
			Abil:           "Burning Shadowhunt Shot (A4)",
			AttackTag:      attacks.AttackTagExtra,
			AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
			ICDTag:         attacks.ICDTagShadowhuntShell,
			ICDGroup:       attacks.ICDGroupShadowhuntShell,
			StrikeType:     attacks.StrikeTypeDefault,
			Mult:           shadowhunt[c.TalentLvlSkill()] * 1.5,
			Element:        attributes.Anemo,
			Durability:     25,
		}

		if c.typeCount >= 1 {
			ai.Abil = "Converted Burning Shadowhunt Shot (A4)"
			ai.Mult = shiningshadowhunt[c.TalentLvlSkill()] * 1.5
			ai.ICDTag = attacks.ICDTagNone
			ai.ICDGroup = attacks.ICDGroupDefault
		}

		c.Core.QueueAttack(
			ai,
			combat.NewSingleTargetHit(c.Core.Combat.PrimaryTarget().Key()),
			37,
			37,
		)
		return false
	}, "chasca-a4")
}
