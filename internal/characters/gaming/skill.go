package gaming

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
)

var skillFrames []int

const (
	skillCD      = 6 * 60
	skillCDStart = 4
)

func init() {
	skillFrames = frames.InitAbilSlice(64)
	skillFrames[action.ActionLowPlunge] = 23
	skillFrames[action.ActionHighPlunge] = 23
	skillFrames[action.ActionJump] = 63
	skillFrames[action.ActionWalk] = 62
	skillFrames[action.ActionSwap] = 63
}

// TODO: currently assuming skill always hits and it hits the earliest possible
// additional delay is user controlled
func (c *char) Skill(p map[string]int) (action.Info, error) {
	c.SetCDWithDelay(action.ActionSkill, skillCD, skillCDStart)
	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionLowPlunge], // earliest cancel
		State:           action.SkillState,
	}, nil
}
