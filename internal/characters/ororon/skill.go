package ororon

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/targets"
)

var skillFrames []int

const (
	skillKey      = "ororon-skill-a1-hook"
	skillHitmark  = 24
	skillInterval = 70
)

func init() {
	skillFrames = frames.InitAbilSlice(12) // E -> E
}

func (c *char) Skill(p map[string]int) (action.Info, error) {

	c.AddStatus(skillKey, 15*60, true)

	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Night's Sling: Spirit Orb DMG (E)",
		AttackTag:      attacks.AttackTagElementalArt,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagElementalArt,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypePierce,
		Element:        attributes.Electro,
		Durability:     25,
		Mult:           skill[c.TalentLvlSkill()],
	}

	skillhits := 4
	if c.Base.Cons >= 1 {
		skillhits = 6
	}
	area := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 10)
	enemies := c.Core.Combat.RandomEnemiesWithinArea(area, nil, skillhits)

	j := 0
	for _, e := range enemies {
		c.Core.QueueAttack(
			ai,
			combat.NewSingleTargetHit(e.Key()),
			skillHitmark+skillInterval*j,
			skillHitmark+skillInterval*j,
			c.a4CB,
			c.c1cb,
			c.particleCB,
		)
		j++
	}

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
	count := 3.0
	c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Electro, c.ParticleDelay)
}
