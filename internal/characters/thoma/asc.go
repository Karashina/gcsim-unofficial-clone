package thoma

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
)

// 現在のアクティブキャラクターが烈炎の障壁を取得または更新した時、
// そのキャラクターのシールド強度が6秒間5%上昇する。
// この効果は0.3秒ごとに1回発動可能。最大5重。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	c.Core.Player.Shields.AddShieldBonusMod("thoma-a1", -1, func() (float64, bool) {
		if c.Tags["shielded"] == 0 {
			return 0, false
		}
		if !c.StatusIsActive("thoma-a1") {
			return 0, false
		}
		return float64(c.a1Stack) * 0.05, true
	})

	c.Core.Events.Subscribe(event.OnCharacterSwap, func(_ ...interface{}) bool {
		c.a1Stack = 0
		return false
	}, "thoma-a1-swap")
}

// 真紅の炎槌の炎崩壊によるダメージがトーマのHP上限の2.2%分上昇する。
func (c *char) a4() float64 {
	if c.Base.Ascension < 1 {
		return 0
	}
	return 0.022 * c.MaxHP()
}
