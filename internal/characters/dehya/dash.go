package dehya

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
)

const (
	burstDashDuration      = 7 // 元素爆発中にジャンプキャンセルした場合の最小ダッシュ時間の推定値
	jumpKickWindowKey      = "dehya-jump-kick-window"
	jumpKickWindowDuration = 17 // 推定値
)

func (c *char) Dash(p map[string]int) (action.Info, error) {
	// フレーム数を決定
	length := c.DashLength()
	canQueueAfter := length
	// 元素爆発がアクティブならフレーム数を調整する必要がある
	if c.StatusIsActive(burstKey) {
		canQueueAfter = burstDashDuration
		// ジャンプでディヘヤがキックに遷移するウィンドウのステータスを追加
		c.AddStatus(jumpKickWindowKey, jumpKickWindowDuration, false)
	}

	// スタミナ処理のデフォルト実装を呼び出す
	c.Character.Dash(p)

	return action.Info{
		Frames: func(next action.Action) int {
			if c.StatusIsActive(burstKey) && next == action.ActionJump {
				return canQueueAfter
			}
			return length
		},
		AnimationLength: length,
		CanQueueAfter:   canQueueAfter,
		State:           action.DashState,
	}, nil
}
