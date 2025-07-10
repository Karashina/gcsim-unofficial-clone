package skirk

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/targets"
)

var (
	skillFrames     []int
	skillHoldFrames []int
)

const (
	particleICDKey = "skirk-particle-icd"
	particleICD    = 9999 * 60

	skillDelay = 26
)

func init() {
	skillFrames = frames.InitAbilSlice(27)
	skillFrames[action.ActionAttack] = 26

	skillHoldFrames = frames.InitAbilSlice(27)
	skillHoldFrames[action.ActionAttack] = 26
	skillHoldFrames[action.ActionDash] = 5
}

func (c *char) reduceSerpentsSubtlety(val float64) {
	c.serpentsSubtlety -= val
	if c.serpentsSubtlety <= 0.00001 {
		c.cancelSevenPhaseFlash()
	}
	c.Core.Log.NewEvent("serpents subtlety reduced", glog.LogEnergyEvent, c.Index).
		Write("reduced", val).Write("current point", c.serpentsSubtlety)
}

func (c *char) generateSerpentsSubtlety(val float64) {
	c.serpentsSubtlety += val
	if c.serpentsSubtlety >= c.serpentsSubtletyMax {
		c.serpentsSubtlety = c.serpentsSubtletyMax
	}
	c.Core.Log.NewEvent("serpents subtlety generated", glog.LogEnergyEvent, c.Index).
		Write("generated", val).Write("current point", c.serpentsSubtlety)
}

func (c *char) cancelSevenPhaseFlash() {
	if !c.onSevenPhaseFlash {
		return
	}
	c.onSevenPhaseFlash = false
	c.DeleteStatus(burstkey)
	c.DeleteStatMod("skirk-c2")
	c.serpentsSubtlety = 0
	c.SetCD(action.ActionSkill, 8*60)
}

func (c *char) serpentsSubtletyReduceFunc(src int) func() {
	return func() {
		if c.sevenPhaseFlashsrc != src {
			return
		}

		if !c.onSevenPhaseFlash {
			return
		}

		c.reduceSerpentsSubtlety(0.7)

		// reduce 1 point per 6f
		c.QueueCharTask(c.serpentsSubtletyReduceFunc(src), 6)
	}
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	hold := p["hold"]
	if hold > 0 {
		return c.SkillHold(p)
	}
	c.QueueCharTask(func() {
		c.onSevenPhaseFlash = true
		if c.Base.Cons < 2 {
			c.generateSerpentsSubtlety(45)
		} else {
			c.generateSerpentsSubtlety(55)
		}
		c.DeleteStatus(particleICDKey)
		c.c2()
		c.sevenPhaseFlashsrc = c.Core.F
		c.QueueCharTask(func() {
			c.serpentsSubtletyReduceFunc(c.sevenPhaseFlashsrc)()
		}, 6)
		c.QueueCharTask(func() {
			c.cancelSevenPhaseFlash()
		}, 12*60)
	}, skillDelay)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionAttack],
		State:           action.SkillState,
	}, nil
}

func (c *char) SkillHold(p map[string]int) (action.Info, error) {

	c.a1(false)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillHoldFrames),
		AnimationLength: skillHoldFrames[action.InvalidAction],
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

	count := 4.0
	c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Cryo, c.ParticleDelay)
}

func (c *char) onExitField() {
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(_ ...interface{}) bool {
		if c.onSevenPhaseFlash {
			c.cancelSevenPhaseFlash()
			c.serpentsSubtlety = 0
		}
		return false
	}, "skirk-exit")
}
