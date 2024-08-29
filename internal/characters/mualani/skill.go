package mualani

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/targets"
)

const (
	skillKey       = "Nightsoul's Blessing: Mualani"
	skillMarkKey   = "mualani-mark"
	markICDKey     = "mualani-mark-icd"
	NAICDKey       = "mualani-NA-icd"
	particleICDKey = "mualani-particle-icd"
)

var skillFramesNormal []int
var skillFramesEnd []int

func init() {
	skillFramesNormal = frames.InitAbilSlice(16)
	skillFramesEnd = frames.InitAbilSlice(25)
}

func (c *char) skillActivate() action.Info {
	c.AddNightsoul("kachina-skill-init", 60)
	c.AddStatus(skillKey, -1, true)
	c.pufferCount = 0
	c.OnNightsoul = true
	c.Core.Tasks.Add(c.depleteNightsoulPoints, 6)
	c.skillMarkTargets()
	c.c2()

	// Return ActionInfo
	return action.Info{
		Frames:          frames.NewAbilFunc(skillFramesNormal),
		AnimationLength: skillFramesNormal[action.InvalidAction],
		CanQueueAfter:   skillFramesNormal[action.ActionSwap], // earliest cancel
		State:           action.SkillState,
	}
}

func (c *char) skillDeactivate() action.Info {
	c.skillEndRoutine()
	delay := 25
	c.SetCD(action.ActionSkill, 6*60)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFramesEnd),
		AnimationLength: delay,
		CanQueueAfter:   delay,
		State:           action.Idle,
	}
}

func (c *char) skillEndRoutine() {
	c.DeleteStatus(skillKey)
	c.DeleteStatus(particleICDKey)
	c.Core.Player.SwapCD = 25
	c.NightsoulPoint = 0
	c.c1count = 0
	c.OnNightsoul = false
}

func (c *char) depleteNightsoulPoints() {
	if c.StatusIsActive(skillKey) {
		c.ConsumeNightsoul(1)
		c.Core.Tasks.Add(c.depleteNightsoulPoints, 6)
	}
	if c.NightsoulPoint <= 0 {
		c.skillEndRoutine()
	}
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	if !c.StatusIsActive(skillKey) {
		return c.skillActivate(), nil
	}
	return c.skillDeactivate(), nil
}

func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if !c.StatusIsActive(skillKey) {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, -1, true)

	count := 3.0
	if c.Core.Rand.Float64() < .66 {
		count = 4
	}
	c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Hydro, c.ParticleDelay)
}

func (c *char) skillMarkTargets() {
	if !c.StatusIsActive(skillKey) {
		return
	}
	enemycount := 0
	area := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{X: 0.0, Y: 0.0}, 2)
	enemies := c.Core.Combat.EnemiesWithinArea(area, nil)
	for _, e := range enemies {
		if enemycount < 5 {
			if !e.StatusIsActive(markICDKey) {
				e.AddStatus(skillMarkKey, 600, true)
				e.AddStatus(markICDKey, 0.7*60, true)
				enemycount++
				c.WaveMomentum++
				c.Core.Log.NewEvent("wave momentum added", glog.LogCharacterEvent, c.Index).
					Write("stacks:", c.WaveMomentum)
			}
		}
	}
	c.Core.Tasks.Add(c.skillMarkTargets, 10)
}
