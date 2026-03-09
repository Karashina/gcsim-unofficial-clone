package wanderer

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
)

var (
	a4Release   = []int{16, 18, 21, 25}
	dashFramesE []int
)

const a4Hitmark = 30

func init() {
	dashFramesE = frames.InitAbilSlice(24)
	dashFramesE[action.ActionAttack] = 21
	dashFramesE[action.ActionCharge] = 21
	dashFramesE[action.ActionSkill] = 5
	dashFramesE[action.ActionDash] = 22
	dashFramesE[action.ActionJump] = 22
	dashFramesE[action.ActionWalk] = 22
}

func (c *char) Dash(p map[string]int) (action.Info, error) {
	delay := c.checkForSkillEnd()

	if c.StatusIsActive(skillKey) {
		return c.WindfavoredDash(p)
	}

	// 遅延後までダッシュCDとスタミナ消費を延期（遅延は落下をシミュレート）
	c.Core.Tasks.Add(func() {
		c.ApplyDashCD()
		c.QueueDashStaminaConsumption(p)
	}, delay)

	// 長さは標準ダッシュ長 + スキル終了遅延（落下をシミュレート）
	length := c.DashLength() + delay
	return action.Info{
		Frames:          func(action.Action) int { return length },
		AnimationLength: length,
		CanQueueAfter:   length,
		State:           action.DashState,
	}, nil
}

func (c *char) WindfavoredDash(p map[string]int) (action.Info, error) {
	ai := action.Info{
		Frames:          func(next action.Action) int { return dashFramesE[next] },
		AnimationLength: dashFramesE[action.InvalidAction],
		CanQueueAfter:   dashFramesE[action.ActionSkill],
		State:           action.DashState,
	}

	a4Triggered := c.a4()
	if !a4Triggered {
		c.skydwellerPoints -= 15
	}

	return ai, nil
}
