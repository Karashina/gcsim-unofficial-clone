package mualani

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
)

var skillJumpFrames []int

func init() {
	skillJumpFrames = frames.InitAbilSlice(54) // skill
	skillJumpFrames[action.ActionAttack] = 4
	skillJumpFrames[action.ActionBurst] = 50
	skillJumpFrames[action.ActionDash] = 49
	skillJumpFrames[action.ActionJump] = 50
	skillJumpFrames[action.ActionWalk] = 47
	skillJumpFrames[action.ActionSwap] = 48
}

func (c *char) Jump(p map[string]int) (action.Info, error) {
	if c.nightsoulState.HasBlessing() {
		if c.Core.Player.LastAction.Type == action.ActionDash {
			c.reduceNightsoulPoints(14) // 合計24、ダッシュから10、ダッシュジャンプから14
		} else {
			c.reduceNightsoulPoints(2)
		}

		return action.Info{
			Frames:          frames.NewAbilFunc(skillJumpFrames),
			AnimationLength: skillJumpFrames[action.InvalidAction],
			CanQueueAfter:   skillJumpFrames[action.ActionWalk],
			State:           action.JumpState,
		}, nil
	}
	return c.Character.Jump(p)
}
