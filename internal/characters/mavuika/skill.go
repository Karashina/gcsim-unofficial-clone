package mavuika

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/targets"
)

var skillFrames []int
var skillHoldFrames []int
var SkillChangeFrames []int

const (
	bikeKey  = "mavuika-bike"
	skillKey = "mavuika-skill"
)

func init() {
	skillFrames = frames.InitAbilSlice(31)

	skillHoldFrames = frames.InitAbilSlice(43)

	SkillChangeFrames = frames.InitAbilSlice(27)
}

func (c *char) reduceNightsoulPoints(val float64) {
	if c.StatusIsActive(BurstKey) {
		return
	}
	c.nightsoulState.ConsumePoints(val)
	if c.nightsoulState.Points() <= 0.00001 {
		c.cancelNightsoul()
	}
}

func (c *char) enterNightsoul(amt float64, pointdelay int) {
	if !c.nightsoulState.HasBlessing() {
		c.nightsoulState.EnterBlessing(amt)
		c.QueueCharTask(func() {
			c.nightsoulPointReduceFunc(c.nightsoulSrc)()
		}, pointdelay)
	} else {
		c.QueueCharTask(func() {
			c.nightsoulState.GeneratePoints(amt)
		}, pointdelay)
	}
}

func (c *char) cancelNightsoul() {
	c.nightsoulState.ExitBlessing()
	c.DeleteStatus(bikeKey)
	c.nightsoulSrc = -1
	c.c2DefModRemove()
	c.c6DefModRemove()
}

func (c *char) nightsoulPointReduceFunc(src int) func() {
	return func() {
		if c.nightsoulSrc != src {
			return
		}

		if !c.nightsoulState.HasBlessing() {
			return
		}

		if c.StatusIsActive(bikeKey) {
			if c.StatusIsActive(chargeKey) {
				c.reduceNightsoulPoints(0.5 * 2.2)
			} else {
				c.reduceNightsoulPoints(0.5 * 1.8)
			}
		}

		c.reduceNightsoulPoints(0.5)

		// reduce 0.5 point per 6f
		c.QueueCharTask(c.nightsoulPointReduceFunc(src), 6)
	}
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(skillKey) && c.nightsoulState.HasBlessing() && !c.StatusIsActive(BurstKey) {
		//toggle bike
		if c.StatusIsActive(bikeKey) {
			c.DeleteStatus(bikeKey)
			c.QueueCharTask(c.searingradiance, 120)
			c.c6DefModRemove()
			c.c2DefModAdd()
		} else {
			c.AddStatus(bikeKey, -1, false)
			c.c6()
			c.c6DefModAdd()
			c.c2DefModRemove()
		}
		return action.Info{
			Frames:          frames.NewAbilFunc(SkillChangeFrames),
			AnimationLength: SkillChangeFrames[action.InvalidAction],
			CanQueueAfter:   SkillChangeFrames[action.ActionAttack], // earliest cancel
			State:           action.SkillState,
		}, nil
	}

	// Initial Skill Damage
	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "The Named Moment(Press)",
		AttackTag:      attacks.AttackTagElementalArt,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagNone,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Pyro,
		Durability:     25,
		Mult:           skill[c.TalentLvlSkill()],
	}
	canQueueAfter := skillFrames[action.ActionAttack]
	AnimationLength := skillFrames[action.InvalidAction]
	skillHitmark := 13

	hold := max(p["hold"], 0)
	if hold > 0 || c.StatusIsActive(BurstKey) {
		ai.Abil = "The Named Moment(Hold)"
		ai.StrikeType = attacks.StrikeTypeBlunt
		skillHitmark = 28
		c.AddStatus(bikeKey, -1, false)
		canQueueAfter = skillHoldFrames[action.ActionAttack]
		AnimationLength = skillHoldFrames[action.InvalidAction]
		c.c2DefModRemove()
		c.c6DefModAdd()
		c.c6()
	} else {
		c.QueueCharTask(c.searingradiance, 135)
		c.c2DefModAdd()
		c.c6DefModRemove()
	}

	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 6), skillHitmark, skillHitmark, c.particleCB)

	c.AddStatus(skillKey, -1, false)
	c.SetCD(action.ActionSkill, 15*60)
	c.enterNightsoul(80, 36)

	return action.Info{
		Frames: func(next action.Action) int {
			if hold > 0 {
				return skillHoldFrames[next]
			} else {
				return skillFrames[next]
			}
		},
		AnimationLength: AnimationLength,
		CanQueueAfter:   canQueueAfter,
		State:           action.SkillState,
	}, nil
}

func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	count := 5.0
	c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Pyro, c.ParticleDelay)
}

func (c *char) searingradiance() {
	if !c.nightsoulState.HasBlessing() {
		return
	}
	if c.StatusIsActive(bikeKey) {
		return
	}
	c.reduceNightsoulPoints(3)
	// Skill DoT Damage
	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Ring of Searing Radiance DMG",
		AttackTag:      attacks.AttackTagElementalArt,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagNone,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Pyro,
		Durability:     25,
		Mult:           skilldot[c.TalentLvlSkill()],
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
		c.Core.QueueAttack(ai, combat.NewSingleTargetHit(enemy.Key()), 0, 0, c.c6SkillCB())
		remaining--
	}
	for _, gadget := range gadgets {
		if remaining <= 0 {
			break
		}
		c.Core.QueueAttack(ai, combat.NewSingleTargetHit(gadget.Key()), 0, 0)
		remaining--
	}
	c.QueueCharTask(c.searingradiance, 119)
}
