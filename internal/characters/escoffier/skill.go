package escoffier

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
	"github.com/genshinsim/gcsim/pkg/core/targets"
)

const (
	arkheCD        = "skill-arkhe-cd"
	particleICDKey = "skill-particle-icd"

	skillHitmark    = 45
	cookingMekSpawn = 41
	arkheHitmark    = 88
)

var skillFrames []int

func init() {
	skillFrames = frames.InitAbilSlice(38)
	skillFrames[action.ActionAttack] = 38
	skillFrames[action.ActionDash] = 45
}

func (c *char) Skill(p map[string]int) (action.Info, error) {

	player := c.Core.Combat.Player()

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Low-Temperature Cooking (E)",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagElementalArt,
		ICDGroup:   attacks.ICDGroupEscoffierCookingMek,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Cryo,
		Durability: 25,
		Mult:       skill[c.TalentLvlSkill()],
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(player, geometry.Point{Y: 2.6}, 4.5),
		skillHitmark,
		skillHitmark,
		c.particleCB,
	)

	c.QueueCharTask(func() {
		c.spawnCookingMek()
	}, cookingMekSpawn)
	c.arkheAttack()
	c.SetCD(action.ActionSkill, 15*60)
	c.AddStatus(c2Key, 15*60, true)
	c.c2count = 0
	c.c6count = 0

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash], // earliest cancel
		State:           action.SkillState,
	}, nil
}

func (c *char) arkheAttack() {
	if c.StatusIsActive(arkheCD) {
		return
	}
	c.AddStatus(arkheCD, 10*60, true)

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Surging Blade (E-Arkhe)",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Cryo,
		Durability: 0,
		Mult:       skillArkhe[c.TalentLvlSkill()],
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), geometry.Point{Y: 2.6}, 4.5),
		arkheHitmark,
		arkheHitmark,
	)
}

func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	c.Core.QueueParticle(c.Base.Key.String(), 4, attributes.Cryo, c.ParticleDelay)
}
