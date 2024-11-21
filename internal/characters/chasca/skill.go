package chasca

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/targets"
)

var (
	skillFrames       []int
	skillHitmark      = 10
	skillCancelFrames []int
)

const (
	particleICDKey = "chasca-particle-icd"
	momentumICDKey = "chasca-momentum-icd"
	resonatingKey  = "chasca-resonating"

	skillDelay  = 2
	particleICD = 9999 * 60
	momentumIcd = 0.7 * 60
)

func init() {
	skillFrames = frames.InitAbilSlice(19)
	skillFrames[action.ActionAim] = 19

	skillCancelFrames = frames.InitAbilSlice(46)
}

func (c *char) reduceNightsoulPoints(val float64) {
	c.nightsoulState.ConsumePoints(val)
	if c.nightsoulState.Points() <= 0.00001 {
		c.cancelNightsoul()
	}
}

func (c *char) cancelNightsoul() {
	c.nightsoulState.ExitBlessing()
	c.SetCD(action.ActionSkill, 6.5*60)
	c.DeleteStatus(resonatingKey)
	c.nightsoulSrc = -1
}

func (c *char) nightsoulPointReduceFunc(src int) func() {
	return func() {
		if c.nightsoulSrc != src {
			return
		}

		if !c.nightsoulState.HasBlessing() {
			return
		}

		c.reduceNightsoulPoints(0.8)

		// reduce 0.8 point per 8f
		c.QueueCharTask(c.nightsoulPointReduceFunc(src), 6)
	}
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	if c.nightsoulState.HasBlessing() {
		c.cancelNightsoul()
		if p["hold"] == 1 {
			return c.highPlunge(), nil
		}
		return action.Info{
			Frames:          frames.NewAbilFunc(skillCancelFrames),
			AnimationLength: skillCancelFrames[action.InvalidAction],
			CanQueueAfter:   skillCancelFrames[action.ActionAttack], // earliest cancel
			State:           action.SkillState,
		}, nil
	}

	c.QueueCharTask(func() {
		c.nightsoulState.EnterBlessing(80)
		c.DeleteStatus(particleICDKey)
		c.nightsoulSrc = c.Core.F
		c.QueueCharTask(func() {
			c.AddStatus(resonatingKey, -1, false) // enter resonating state
			c.nightsoulPointReduceFunc(c.nightsoulSrc)()
		}, 6)
	}, skillDelay)

	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Resonance DMG (E)",
		AttackTag:      attacks.AttackTagExtra,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagNone,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeDefault,
		Mult:           resonance[c.TalentLvlSkill()],
		Element:        attributes.Anemo,
		Durability:     25,
	}

	c.Core.QueueAttack(
		ai,
		combat.NewSingleTargetHit(c.Core.Combat.PrimaryTarget().Key()),
		skillHitmark,
		skillHitmark,
	)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionAttack],
		State:           action.SkillState,
	}, nil
}

func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, particleICD, true)

	count := 5.0
	c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Anemo, c.ParticleDelay)
}
