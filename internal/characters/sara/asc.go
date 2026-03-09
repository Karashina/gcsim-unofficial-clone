package sara

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

// 固有天賦1は aimed.go で実装済み:
// 天狗召嗚のクロウフェザー保護状態中、狙い撃ちのチャージ時間が60%短縮される。

// 厳密には必要ないが、将来プレイヤーが攻撃を受ける実装をした場合に備えて
const a4ICDKey = "sara-a4-icd"

// 天狗呉雷・待ち伏せが敵に命中すると、九條裟羅の元素チャージ効率100%ごとに
// チーム全員に1.2エネルギーを回復する。この効果は3秒に1回発動可能。
//
// - ライブラリの調査によると、テキスト説明は不正確
//
// - 実際には元素チャージ効率1%ごとに0.012の固定エネルギーを付与
func (c *char) makeA4CB() combat.AttackCBFunc {
	if c.Base.Ascension < 4 {
		return nil
	}
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if c.StatusIsActive(a4ICDKey) {
			return
		}
		c.AddStatus(a4ICDKey, 180, true)

		energyAddAmt := 1.2 * c.NonExtraStat(attributes.ER)
		for _, char := range c.Core.Player.Chars() {
			char.AddEnergy("sara-a4", energyAddAmt)
		}
	}
}
