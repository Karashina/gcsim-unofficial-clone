package clorinde

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
	"github.com/genshinsim/gcsim/pkg/core/targets"
)

var skillFrames []int

func init() {
	skillFrames = frames.InitAbilSlice(34)
}

const (
	skillAlignedICDKey = "clorinde-aligned-icd"
	skillCDKey         = "clorinde-skill-cd"
	skillAlignedICD    = 10 * 60
	skillBuffKey       = "Night Vigil"
)

func (c *char) Skill(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(skillBuffKey) {
		c.ImpaletheNight(p)
	}
	if !c.StatusIsActive(skillBuffKey) && c.StatusIsActive(skillCDKey) {
		return action.Info{
			Frames:          frames.NewAbilFunc(frames.InitAbilSlice(0)),
			AnimationLength: 0,
			CanQueueAfter:   0, // earliest cancel
			State:           action.SkillState,
		}, nil
	}

	// start skill buff on cast
	c.AddStatus(skillBuffKey, 7.5*60, true)
	c.AddStatus(skillCDKey, 16*60, true)

	if c.Base.Cons >= 6 {
		c.c6init()
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash], // earliest cancel
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
	c.AddStatus(particleICDKey, 2*60, true)

	count := 1.0
	c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Electro, c.ParticleDelay) // TODO: this used to be 80 for particle delay
}

func (c *char) onExitField() {
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(args ...interface{}) bool {
		prev := args[0].(int)
		if prev != c.Index {
			return false
		}
		c.DeleteStatus(skillBuffKey)
		// queue up a4
		c.Core.Tasks.Add(c.a4, 60)
		return false
	}, "clorinde-exit")
}

func (c *char) skillAligned() {
	if c.StatusIsActive(skillAlignedICDKey) {
		return
	}
	c.AddStatus(skillAlignedICDKey, skillAlignedICD, true)

	skillAlignedAI := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Surging Blade (" + c.Base.Key.Pretty() + ")",
		AttackTag:          attacks.AttackTagElementalArt,
		ICDTag:             attacks.ICDTagNone,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeSpear,
		Element:            attributes.Electro,
		Durability:         0,
		Mult:               alkhe[c.TalentLvlSkill()],
		HitlagFactor:       0.01,
		CanBeDefenseHalted: true,
	}

	c.Core.QueueAttack(
		skillAlignedAI,
		combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: -0.3}, 1.2, 4.5),
		15,
		15,
	)
}
