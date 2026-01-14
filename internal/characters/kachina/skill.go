package kachina

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
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
		return c.rideTwirly(), nil
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
	c.AddStatus(skillRideKey, -1, true)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillHoldFrames),
		AnimationLength: skillHoldFrames[action.InvalidAction],
		CanQueueAfter:   skillHoldFrames[action.ActionDash], // earliest cancel
		State:           action.SkillState,
	}
}

func (c *char) skillInit() {
	c.nightsoulState.EnterBlessing(60)
	c.AddStatus(skillKey, -1, true)
	c.newTwirly()
}

func (c *char) skillEndRoutine() {
	c.DeleteStatus(skillKey)
	c.DeleteStatus(skillRideKey)
	c.nightsoulState.ExitBlessing()
	c.nightsoulState.ClearPoints()
	c.removeTwirly()
}

func (c *char) rideTwirly() action.Info {
	c.c1shard()
	if c.StatusIsActive(skillRideKey) {
		c.DeleteStatus(skillRideKey)
		c.newTwirly()
	} else {
		c.AddStatus(skillRideKey, -1, true)
	}
	return action.Info{
		Frames:          frames.NewAbilFunc(skillPressFrames),
		AnimationLength: skillPressFrames[action.InvalidAction],
		CanQueueAfter:   skillPressFrames[action.ActionDash], // earliest cancel
		State:           action.SkillState,
	}
}

func (c *char) depleteNightsoulPoints(t string) {
	if !c.StatusIsActive(skillKey) {
		return
	}
	switch t {
	case "attack":
		c.nightsoulState.ConsumePoints(10)
	case "dismount":
		c.nightsoulState.ConsumePoints(2)
	default:
	}
	if c.nightsoulState.Points() <= 0 {
		c.skillEndRoutine()
	}
}

