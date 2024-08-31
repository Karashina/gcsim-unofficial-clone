package kachina

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
)

const (
	skillKey     = "Nightsoul's Blessing: Kachina"
	skillRideKey = "kachina-riding-twirly"
)

var skillPressFrames []int
var skillHoldFrames []int

func init() {
	skillPressFrames = frames.InitAbilSlice(49)
	skillHoldFrames = frames.InitAbilSlice(47)
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(skillKey) {
		c.rideTwirly()
	}
	h := p["hold"]
	if h > 0 {
		return c.skillHold(), nil
	}
	return c.skillPress(), nil
}

func (c *char) skillPress() action.Info {
	c.skillInit()
	c.SetCD(action.ActionSkill, 20*60)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillPressFrames),
		AnimationLength: skillPressFrames[action.InvalidAction],
		CanQueueAfter:   skillPressFrames[action.ActionDash], // earliest cancel
		State:           action.SkillState,
	}
}

func (c *char) skillHold() action.Info {
	c.skillInit()
	c.SetCD(action.ActionSkill, 20*60)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillHoldFrames),
		AnimationLength: skillHoldFrames[action.InvalidAction],
		CanQueueAfter:   skillHoldFrames[action.ActionDash], // earliest cancel
		State:           action.SkillState,
	}
}

func (c *char) skillInit() {
	c.AddNightsoul("kachina-skill-init", 60)
	c.AddStatus(skillKey, -1, true)
	c.OnNightsoul = true
	c.newTwirly()
}

func (c *char) skillEndRoutine() {
	c.DeleteStatus(skillKey)
	c.DeleteStatus(skillRideKey)
	c.NightsoulPoint = 0
	c.OnNightsoul = false
	c.removeTwirly()
}

func (c *char) rideTwirly() {
	c.c1shard()
	if c.StatusIsActive(skillRideKey) {
		c.DeleteStatus(skillRideKey)
		c.newTwirly()
	} else {
		c.AddStatus(skillRideKey, -1, true)
	}
}

func (c *char) depleteNightsoulPoints(t string) {
	if !c.StatusIsActive(skillKey) {
		return
	}
	switch t {
	case "attack":
		c.ConsumeNightsoul(10)
	case "dismount":
		c.ConsumeNightsoul(2)
	default:
	}
	if c.NightsoulPoint <= 0 {
		c.skillEndRoutine()
	}
}