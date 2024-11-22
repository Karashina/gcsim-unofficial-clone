package ororon

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/targets"
)

const (
	a1hitmark        = 15
	a1SkillICDKey    = "ororon-a1-skill-icd"
	a1blessingICDKey = "ororon-a1-blessing-icd"
	a1DurKey         = "ororon-nightsoul-blessing"
	a4Key            = "ororon-a4-active"
	a4ICDKey         = "ororon-a4-icd"
)

func containsTag(tags []attacks.AdditionalTag, tag attacks.AdditionalTag) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}

	c.Core.Events.Subscribe(event.OnNightsoulBurst, func(args ...interface{}) bool {
		c.nightsoulState.GeneratePoints(40)
		return false
	}, "ororon-a1-onnightsoulburst")

	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if c.StatusIsActive(skillKey) && !c.StatusIsActive(a1SkillICDKey) && c.a1Count > 0 && atk.Info.ActorIndex != c.Index {
			if atk.Info.Element == attributes.Hydro || atk.Info.Element == attributes.Electro {
				c.a1Count--
				c.nightsoulState.GeneratePoints(5)
				c.AddStatus(a1SkillICDKey, 0.3*60, true)
			}
		}
		if atk.Info.Abil == "electrocharged" || containsTag(atk.Info.AdditionalTags, attacks.AdditionalTagNightsoul) {
			if c.nightsoulState.Points() >= 10 && !c.StatusIsActive(a1blessingICDKey) {
				c.nightsoulState.EnterBlessing(c.nightsoulState.Points())
				c.nightsoulState.ConsumePoints(10)
				c.AddStatus(a1blessingICDKey, 1.8*60, true)
				c.AddStatus(a1DurKey, 6*60, true)

				ai := combat.AttackInfo{
					ActorIndex:     c.Index,
					Abil:           "Hypersense (A1)",
					AttackTag:      attacks.AttackTagNone,
					AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
					ICDTag:         attacks.ICDTagElementalArt,
					ICDGroup:       attacks.ICDGroupDefault,
					StrikeType:     attacks.StrikeTypePierce,
					Element:        attributes.Electro,
					Durability:     25,
					Mult:           1.6,
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
		}
		return false
	}, "ororon-a1-onhit")

	c.Core.Events.Subscribe(event.OnTick, func(args ...interface{}) bool {
		if !c.nightsoulState.HasBlessing() {
			return false
		}
		if c.StatusIsActive(a1DurKey) {
			return false
		}
		c.nightsoulState.ExitBlessing()
		return false
	}, "ororon-a1-ontick")
}

func (c *char) a4CB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	c.AddStatus(a4Key, 15*60, true)
}

func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if !c.StatusIsActive(a4Key) {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagNormal && atk.Info.AttackTag != attacks.AttackTagExtra && atk.Info.AttackTag != attacks.AttackTagPlunge {
			return false
		}

		c.Core.Player.ByIndex(atk.Info.ActorIndex).AddEnergy("ororon-a4", 3)
		if c.Core.Player.ActiveChar().Index != c.Index {
			c.AddEnergy("ororon-a4-self", 3)
		}
		c.AddStatus(a4ICDKey, 60, true)
		c.a4Count--

		return false
	}, "ororon-a4-onhit")
}
