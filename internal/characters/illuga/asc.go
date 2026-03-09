package illuga

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	a1CritKey = "illuga-a1-crit"
	a1EmKey   = "illuga-a1-em"
	a1Dur     = 20 * 60 // 20s (I-5 fix: was 12s)
)

// 固有天賦1: 灯守の誓い
// 元素スキルまたは元素爆発使用時：
// - 他のパーティメンバーの岩元素ダメージ：会心率+5%、会心ダメージ+10%
// - 月相がAscendant Gleamの場合、影響を受けるパーティメンバーの元素熟知+50
// 持続時間: 20秒

func (c *char) applyLightkeeperOath() {
	if c.Base.Ascension < 1 {
		return
	}

	// 月相がAscendant Gleamかチェック
	isAscendant := c.checkAscendantGleam()

	// 基本ボーナス
	critRateBonus := 0.05
	critDmgBonus := 0.10
	emBonus := 50.0

	// I-6修正: 6命の会心ボーナスは常時適用、元素熟知はAscendant時のみ
	if c.Base.Cons >= 6 {
		critRateBonus = 0.10 // 10% (vs 5%)
		critDmgBonus = 0.30  // 30% (vs 10%)
	}
	if c.Base.Cons >= 6 && isAscendant {
		emBonus = 80.0 // 80 (vs 50)
	}

	// 全パーティメンバーに適用（一貫性のため自身も含む）
	for _, char := range c.Core.Player.Chars() {
		if char.Index == c.Index {
			continue // 自身には適用しない、他のパーティメンバーのみ
		}

		// 岩元素の会心ボーナスを追加
		m := make([]float64, attributes.EndStatType)
		m[attributes.CR] = critRateBonus
		m[attributes.CD] = critDmgBonus

		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBaseWithHitlag(a1CritKey, a1Dur),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				if atk.Info.Element != attributes.Geo {
					return nil, false
				}
				return m, true
			},
		})

		// Ascendant Gleamの場合、元素熟知ボーナスを追加
		if isAscendant {
			mEM := make([]float64, attributes.EndStatType)
			mEM[attributes.EM] = emBonus

			char.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag(a1EmKey, a1Dur),
				AffectedStat: attributes.EM,
				Amount: func() ([]float64, bool) {
					return mEM, true
				},
			})
		}
	}

	c.Core.Log.NewEvent("Illuga A1: Lightkeeper's Oath applied to party", glog.LogCharacterEvent, c.Index).
		Write("crit_rate_bonus", critRateBonus).
		Write("crit_dmg_bonus", critDmgBonus).
		Write("em_bonus", emBonus).
		Write("is_ascendant", isAscendant)
}

// 固有天賦4: 強化されたナイチンゲールの歌
// パーティ構成に基づいてナイチンゲールの歌のボーナスが強化される：
// 注: Oriole-Song計算のmodifierとして実装

func (c *char) a4Init() {
	if c.Base.Ascension < 4 {
		return
	}

	// 水元素と岩元素のパーティメンバーを数える（仕様に従い自身も含む）
	c.a4HydroCount = 0
	c.a4GeoCount = 0

	for _, char := range c.Core.Player.Chars() {
		switch char.Base.Element {
		case attributes.Hydro:
			c.a4HydroCount++
		case attributes.Geo:
			c.a4GeoCount++
		}
	}

	c.Core.Log.NewEvent("Illuga A4: Party composition counted", glog.LogCharacterEvent, c.Index).
		Write("hydro_count", c.a4HydroCount).
		Write("geo_count", c.a4GeoCount)
}

// getA4GeoBonus は固有天賦4のナイチンゲールの歌の岩元素ダメージボーナスをヒット毎に返す。
// パーティに水元素または岩元素キャラクターが1/2/3人いる場合、
// 増加量はイルーガの元素熟知の7%/14%/24%に等しい。
func (c *char) getA4GeoBonus() float64 {
	if c.Base.Ascension < 4 {
		return 0
	}
	count := c.a4HydroCount + c.a4GeoCount
	if count <= 0 {
		return 0
	}
	if count > 3 {
		count = 3
	}
	em := c.Stat(attributes.EM)
	return a4GeoEM[count-1] * em
}

// getA4LCrsBonus は固有天賦4のナイチンゲールの歌のLCrsダメージボーナスをヒット毎に返す。
// パーティに水元素または岩元素キャラクターが1/2/3人いる場合、
// 増加量はイルーガの元素熟知の48%/96%/160%に等しい。
func (c *char) getA4LCrsBonus() float64 {
	if c.Base.Ascension < 4 {
		return 0
	}
	count := c.a4HydroCount + c.a4GeoCount
	if count <= 0 {
		return 0
	}
	if count > 3 {
		count = 3
	}
	em := c.Stat(attributes.EM)
	return a4LCrsEM[count-1] * em
}

// checkAscendantGleam は現在の月相がAscendant Gleamかチェックする
func (c *char) checkAscendantGleam() bool {
	// 初期化時に設定されたパーティ全体の月相ステータスをチェック
	return c.MoonsignAscendant
}
