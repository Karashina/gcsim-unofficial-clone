package mavuika

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/core/targets"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const c6SkillIcdKey = "mavuika-c6-icd-skill"

func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.ATKP] = 0.4
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag("mavuika-c1", 8*60),
		AffectedStat: attributes.ATKP,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})
}

func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}
	c.BaseStats[attributes.BaseATK] = c.BaseStats[attributes.BaseATK] + 200

	c.c2trg = c.Core.Combat.RandomEnemiesWithinArea(combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 7), nil, 10)
}

func (c *char) c2DefModAdd() {
	if c.Base.Cons < 2 {
		return
	}
	for _, e := range c.c2trg {
		e.AddDefMod(combat.DefMod{
			Base:  modifier.NewBaseWithHitlag("mavuika-c2", -1),
			Value: -0.2,
		})
	}
}

func (c *char) c2DefModRemove() {
	if c.Base.Cons < 2 {
		return
	}
	for _, e := range c.c2trg {
		e.DeleteDefMod("mavuika-c2")
	}
}

func (c *char) c6SkillCB() combat.AttackCBFunc {
	if c.Base.Cons < 6 {
		return nil
	}
	if c.StatusIsActive(bikeKey) {
		return nil
	}
	if c.StatusIsActive(c6SkillIcdKey) {
		return nil
	}
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		ai := combat.AttackInfo{
			ActorIndex:     c.Index,
			Abil:           "Flamestrider Crash(C6)",
			AttackTag:      attacks.AttackTagNone,
			AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
			ICDTag:         attacks.ICDTagNone,
			ICDGroup:       attacks.ICDGroupDefault,
			StrikeType:     attacks.StrikeTypeBlunt,
			Element:        attributes.Pyro,
			Durability:     25,
			Mult:           2,
		}
		c.AddStatus(c6SkillIcdKey, 30, false)
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 0, 0)
	}
}

func (c *char) c6() {
	if c.Base.Cons < 6 {
		return
	}
	if !c.nightsoulState.HasBlessing() {
		return
	}
	if !c.StatusIsActive(bikeKey) {
		return
	}
	// Skill DoT Damage
	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Scorching Ring of Searing Radiance(C6)",
		AttackTag:      attacks.AttackTagNone,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagNone,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Pyro,
		Durability:     25,
		Mult:           5,
	}

	enemies := c.Core.Combat.RandomEnemiesWithinArea(combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 7), nil, 10)
	enemyCount := len(enemies)
	gadgets := c.Core.Combat.RandomGadgetsWithinArea(combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 7), nil, 10)
	gadgetCount := len(gadgets)
	totalEntities := enemyCount + gadgetCount

	remaining := min(10, totalEntities)
	for _, enemy := range enemies {
		if remaining <= 0 {
			break
		}
		c.Core.QueueAttack(ai, combat.NewSingleTargetHit(enemy.Key()), 0, 0)
		remaining--
	}
	for _, gadget := range gadgets {
		if remaining <= 0 {
			break
		}
		c.Core.QueueAttack(ai, combat.NewSingleTargetHit(gadget.Key()), 0, 0)
		remaining--
	}
	c.QueueCharTask(c.c6, 180)
}

func (c *char) c6DefModAdd() {
	if c.Base.Cons < 6 {
		return
	}
	for _, e := range c.c2trg {
		e.AddDefMod(combat.DefMod{
			Base:  modifier.NewBaseWithHitlag("mavuika-c6", -1),
			Value: -0.2,
		})
	}
}

func (c *char) c6DefModRemove() {
	if c.Base.Cons < 6 {
		return
	}
	for _, e := range c.c2trg {
		e.DeleteDefMod("mavuika-c6")
	}
}
