package dehya

import (
	"errors"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
)

func (c *char) Jump(p map[string]int) (action.Info, error) {
	// ジャンプキックウィンドウ中のジャンプでキックが実行される
	if c.StatusIsActive(burstKey) && c.StatusIsActive(jumpKickWindowKey) {
		c.burstHitSrc++                   // 他のパンチ/キックタスクを無効化
		c.DeleteStatus(jumpKickWindowKey) // ウィンドウを削除
		return c.burstKick(c.burstHitSrc), nil
	}

	// キック中のジャンプは許可されない
	// TODO: これが実際に発生し得るか不明…
	if c.StatusIsActive(kickKey) {
		return action.Info{}, errors.New("can't jump cancel burst kick")
	}

	// ジャンプ時に元素爆発がアクティブで、ジャンプキックウィンドウ中でなかった場合
	if c.StatusIsActive(burstKey) {
		c.burstHitSrc = -1       // 他のパンチ/キックタスクを無効化
		c.DeleteStatus(burstKey) // 元素爆発を削除
		// 領域を設置
		if dur := c.sanctumSavedDur; dur > 0 {
			c.sanctumSavedDur = 0
			c.addField(dur)
		}
	}

	return c.Character.Jump(p)
}
