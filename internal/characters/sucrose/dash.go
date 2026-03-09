package sucrose

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
)

var dashFrames []int

func init() {
	dashFrames = frames.InitAbilSlice(24)
	dashFrames[action.ActionSkill] = 1
	dashFrames[action.ActionBurst] = 1
}

// スクロースのダッシュはEとQでキャンセル可能なので、ここでオーバーライドする
func (c *char) Dash(p map[string]int) (action.Info, error) {
	// スタミナ処理のデフォルト実装を呼び出す
	c.Character.Dash(p)
	return action.Info{
		Frames:          frames.NewAbilFunc(dashFrames),
		AnimationLength: dashFrames[action.InvalidAction],
		CanQueueAfter:   dashFrames[action.ActionSkill], // 最速キャンセル
		State:           action.DashState,
	}, nil
}
