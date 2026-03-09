package zibai

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/reactions"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// a0Init は月相の加護パッシブを初期化する
// パーティメンバーが水元素結晶反応をトリガーすると、Lunar-Crystallize反応に変換され、
// ジバイの防御力100ごとにLunar-Crystallizeの基礎ダメージが0.7%増加する（最大14%）。
// さらに、ジバイがパーティにいると、パーティの月相レベルが1増加する。
func (c *char) a0Init() {
	// 全パーティメンバーにLCrsキーを付与（Lunar-Crystallizeを有効化）
	for _, char := range c.Core.Player.Chars() {
		char.AddStatus("LCrs-key", -1, true)
	}
	// ジバイにmoonsignKeyを付与（パーティの月相レベルを1増加）
	// これはパーティ初期化時に月相状態を決定するためにカウントされる
	c.AddStatus("moonsignKey", -1, true)
	// 防御力に基づくLunar-Crystallize基礎ダメージボーナスを追加
	// STUB: Lunar-Crystallize反応ダメージ計算を修正する必要あり
	// 防御力100ごとに0.7%ボーナス、最大14%（防御力2000）
	c.AddLCrsBaseReactBonusMod(character.LCrsBaseReactBonusMod{
		Base: modifier.NewBase("the-coursing-sun-and-moon-a0", -1),
		Amount: func(ai combat.AttackInfo) (float64, bool) {
			maxval := 0.14
			val := min(maxval, c.TotalDef(false)/100*0.007)
			return val, false
		},
	})
}

// a1Init は降り立つ仙術パッシブを初期化する
// 元素スキルを発動するか、近くのパーティメンバーがLunar-Crystallize反応ダメージをトリガーすると、
// ジバイがSelenic Descent効果を4秒間獲得する: 神馬駆けの2段目のダメージが
// ジバイの防御力の60%分増加する。
func (c *char) a1Init() {
	if c.Base.Ascension < 1 {
		return
	}
	const selenicDescentDuration = 4 * 60 // 4 seconds

	// 元素スキル発動を購読
	c.Core.Events.Subscribe(event.OnSkill, func(args ...interface{}) bool {
		if c.Core.Player.Active() != c.Index {
			return false
		}
		c.applySelenicDescent(selenicDescentDuration)
		return false
	}, "zibai-a1-skill")

	// パーティメンバーのLunar-Crystallize反応ダメージを購読
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		ae, ok := args[1].(*combat.AttackEvent)
		if !ok {
			return false
		}
		// Lunar-Crystallize反応ダメージのみトリガー
		if ae.Info.Abil != string(reactions.LunarCrystallize) {
			return false
		}
		c.applySelenicDescent(selenicDescentDuration)
		return false
	}, "zibai-a1-lcrs")
}

// applySelenicDescent はSelenic Descentバフを適用する
func (c *char) applySelenicDescent(duration int) {
	c.AddStatus(selenicDescentKey, duration, true)

	c.Core.Log.NewEvent("Zibai gains Selenic Descent", glog.LogCharacterEvent, c.Index).
		Write("duration", duration)
}

// a4Init は層峰雲を突くパッシブを初期化する
// 他の岩元素パーティメンバーがジバイの防御力を15%ずつ増加させる。
// 水元素パーティメンバーが元素熟知だ60ずつ増加させる。
func (c *char) a4Init() {
	if c.Base.Ascension < 4 {
		return
	}

	geoCount := 0
	hydroCount := 0

	for _, char := range c.Core.Player.Chars() {
		if char.Index == c.Index {
			continue // 自分をスキップ
		}
		switch char.Base.Element {
		case attributes.Geo:
			geoCount++
		case attributes.Hydro:
			hydroCount++
		}
	}

	defBonus := 0.15 * float64(geoCount)
	emBonus := 60.0 * float64(hydroCount)

	if defBonus > 0 {
		defMod := make([]float64, attributes.EndStatType)
		defMod[attributes.DEFP] = defBonus
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("zibai-a4-defp", -1),
			AffectedStat: attributes.DEFP,
			Amount: func() ([]float64, bool) {
				return defMod, true
			},
		})
	}

	if emBonus > 0 {
		emMod := make([]float64, attributes.EndStatType)
		emMod[attributes.EM] = emBonus
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("zibai-a4-em", -1),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				return emMod, true
			},
		})
	}

	if defBonus > 0 || emBonus > 0 {
		c.Core.Log.NewEvent("Zibai A4 stat bonus applied", glog.LogCharacterEvent, c.Index).
			Write("geo_count", geoCount).
			Write("hydro_count", hydroCount).
			Write("def_bonus", defBonus).
			Write("em_bonus", emBonus)
	}
}
