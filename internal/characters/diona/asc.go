package diona

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/shield"
)

// 猫爪シールドの保護を受けたキャラクターは移動速度が10%上昇し、スタミナ消費が10%減少する。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	c.Core.Player.AddStamPercentMod("diona-a1", -1, func(_ action.Action) (float64, bool) {
		if c.Core.Player.Shields.Get(shield.DionaSkill) != nil {
			return -0.1, false
		}
		return 0, false
	})
}

// 固有天賦2は未実装:
// TODO: シグネチャーミックスのAoEに入った敵は15秒間攻撃力が10%減少する。
