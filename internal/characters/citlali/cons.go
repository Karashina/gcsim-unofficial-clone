package citlali

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/core/targets"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const c4IcdKey = "citlali-c4-icd"

func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)

		if !c.StatusIsActive(SkillKey) {
			return false
		}

		if c.Tags[SkillKey] <= 0 {
			return false
		}

		if ae.Info.ActorIndex == c.Index {
			return false
		}

		switch ae.Info.AttackTag {
		case attacks.AttackTagNormal:
		case attacks.AttackTagExtra:
		case attacks.AttackTagPlunge:
		case attacks.AttackTagElementalArt:
		case attacks.AttackTagElementalArtHold:
		case attacks.AttackTagElementalBurst:
		default:
			return false
		}

		dmgAdded := 2 * c.Stat(attributes.EM)
		ae.Info.FlatDmg += dmgAdded

		c.Tags[SkillKey] -= 1

		c.Core.Log.NewEvent("citlali c1 adding damage", glog.LogPreDamageMod, ae.Info.ActorIndex).
			Write("damage_added", dmgAdded)

		return false
	}, "citlali-c1")
}

func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.EM] = 125
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag("citlali-c2-self", -1),
		AffectedStat: attributes.EM,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})
	n := make([]float64, attributes.EndStatType)
	n[attributes.EM] = 250
	for _, char := range c.Core.Player.Chars() {
		this := char
		this.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("citlali-c2-buff", -1),
			Amount: func(_ *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
				// char must be active
				if c.Core.Player.Active() != this.Index {
					return nil, false
				}
				if !c.StatusIsActive(SkillKey) {
					return nil, false
				}
				return n, true
			},
		})
	}
}

func (c *char) c4CB(a combat.AttackCB) {
	if c.Base.Cons < 4 {
		return
	}
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(c4IcdKey) {
		return
	}
	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Death Defier's Spirit Skull: Obsidian Spiritvessel Skull DMG",
		AttackTag:      attacks.AttackTagNone,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagElementalBurst,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Cryo,
		Durability:     25,
		FlatDmg:        c.Stat(attributes.EM) * 18,
	}
	enemies := c.Core.Combat.RandomEnemiesWithinArea(combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 6.5), nil, 1)
	for _, enemy := range enemies {
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(enemy.Pos(), nil, 3.5), skullHitmark, skullHitmark)
	}

	c.nightsoulState.GeneratePoints(16)
	c.AddEnergy("citlali-c4", 8)

	c.AddStatus(c4IcdKey, 8*60, true)
}

func (c *char) c6() {
	if c.Base.Cons < 6 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	for _, char := range c.Core.Player.Chars() {
		this := char
		this.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("citlali-c6", -1),
			Amount: func(_ *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
				m[attributes.PyroP] = 0.015 * c.c6count
				m[attributes.HydroP] = 0.015 * c.c6count
				return m, true
			},
		})
	}
	n := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("citlali-c6-self", -1),
		Amount: func(_ *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
			n[attributes.DmgP] = 0.025 * c.c6count
			return m, true
		},
	})
}
