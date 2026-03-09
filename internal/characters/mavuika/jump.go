package mavuika

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player"
)

var bikeJumpFrames []int

func init() {
	bikeJumpFrames = frames.InitAbilSlice(42)
	bikeJumpFrames[action.ActionLowPlunge] = 16
}

func (c *char) Jump(p map[string]int) (action.Info, error) {
	if !c.StatusIsActive(player.XianyunAirborneBuff) && c.armamentState == bike &&
		c.nightsoulState.HasBlessing() && c.Core.Player.CurrentState() == action.WalkState {
		c.canBikePlunge = true
		// ジャンプ持続時間後に落下攻撃の可否をfalseに設定
		// 落下攻撃のアクションは十分長いのでsrcFチェックは不要
		c.QueueCharTask(func() {
			c.canBikePlunge = false
		}, bikeJumpFrames[action.InvalidAction])
		return action.Info{
			Frames:          frames.NewAbilFunc(bikeJumpFrames),
			AnimationLength: bikeJumpFrames[action.InvalidAction],
			CanQueueAfter:   bikeJumpFrames[action.ActionLowPlunge], // 最速キャンセル
			State:           action.JumpState,
		}, nil
	}
	return c.Character.Jump(p)
}
