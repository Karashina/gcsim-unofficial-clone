package varesa

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/targets"
	"github.com/genshinsim/gcsim/pkg/enemy"
)

var skillFrames []int

const (
	particleICDKey      = "varesa-particle-icd"
	skillKey            = "follow-up-strike"
	skillHitmark        = 6
	skillTargetCountTag = "marked"
	skillMarkedTag      = "varesa-skill-marked"
	skillCd             = 9 * 60
)

func init() {
	skillFrames = frames.InitAbilSlice(39) //E -> NA
	skillFrames[action.ActionJump] = 39    //E -> J
	skillFrames[action.ActionCharge] = 20  //E -> J
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	nofp := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Riding the Night-Rainbow",
		AttackTag:      attacks.AttackTagElementalArt,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagVaresaCombatCycle,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Electro,
		Durability:     25,
		Mult:           skill[c.TalentLvlSkill()],
	}

	hasfp := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Riding the Night-Rainbow (Fiery Passion)",
		AttackTag:      attacks.AttackTagElementalArt,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagVaresaCombatCycle,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Electro,
		Durability:     25,
		Mult:           skillfp[c.TalentLvlSkill()],
	}

	ai := nofp
	if c.StatusIsActive(fieryPassionKey) {
		ai = hasfp
	}

	// clear all existing tags
	for _, t := range c.Core.Combat.Enemies() {
		if e, ok := t.(*enemy.Enemy); ok {
			e.SetTag(skillMarkedTag, 0)
		}
	}

	// add a task to loop through targets and mark them
	marked, ok := p[skillTargetCountTag]
	// default 1
	if !ok {
		marked = 1
	}
	c.Core.Tasks.Add(func() {
		for _, t := range c.Core.Combat.Enemies() {
			if marked == 0 {
				break
			}
			e, ok := t.(*enemy.Enemy)
			if !ok {
				continue
			}
			e.SetTag(skillMarkedTag, 1)
			c.Core.Log.NewEvent("Valesa Tackle Hit", glog.LogCharacterEvent, c.Index).
				Write("target", e.Key())
			marked--
		}
	}, skillHitmark)

	c.Core.Tasks.Add(func() {
		for _, t := range c.Core.Combat.Enemies() {
			e, ok := t.(*enemy.Enemy)
			if !ok {
				continue
			}
			if e.GetTag(skillMarkedTag) == 0 {
				continue
			}
			e.SetTag(skillMarkedTag, 0)
			c.Core.Log.NewEvent("damaging marked target", glog.LogCharacterEvent, c.Index).
				Write("target", e.Key())
			marked--
			c.Core.QueueAttack(ai, combat.NewSingleTargetHit(e.Key()), 1, 1, c.particleCB)
		}

	}, skillHitmark)

	if c.StatusIsActive(freeskillkey) {
		c.DeleteStatus(freeskillkey)
	} else {
		c.SetCDWithDelay(action.ActionSkill, skillCd, skillHitmark-2)
	}

	c.a1()
	c.nightsoulState.GeneratePoints(20)
	c.AddStatus(skillKey, -1, false)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionSwap], // earliest cancel
		State:           action.SkillState,
	}, nil
}

func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.particleGenerated {
		return
	}
	c.particleGenerated = true

	count := 2.0
	if c.Core.Rand.Float64() < 0.5 {
		count = 3
	}
	c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Electro, c.ParticleDelay)
}
