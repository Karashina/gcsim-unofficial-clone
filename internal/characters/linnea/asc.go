package linnea

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// a0Init は月相パッシブを初期化する
// パーティメンバーが水元素結晶反応をトリガーすると、Lunar-Crystallize反応に変換される。
// リンネアの防御力100ごとにLunar-Crystallizeの基礎ダメージが0.7%増加（最大14%）。
// パーティの月相レベルが1増加する。
func (c *char) a0Init() {
	// 全パーティメンバーにLCrsキーを付与（Lunar-Crystallizeを有効化）
	for _, char := range c.Core.Player.Chars() {
		char.AddStatus("LCrs-key", -1, true)
	}
	// moonsignKeyを付与（月相レベル+1）
	c.AddStatus("moonsignKey", -1, true)

	// 防御力に基づくLunar-Crystallize基礎ダメージボーナスを追加
	// 防御力100ごとに0.7%、最大14%
	c.AddLCrsBaseReactBonusMod(character.LCrsBaseReactBonusMod{
		Base: modifier.NewBase("linnea-a0-lcrs-base", -1),
		Amount: func(ai combat.AttackInfo) (float64, bool) {
			maxVal := 0.14
			val := min(maxVal, c.TotalDef(false)/100*0.007)
			return val, false
		},
	})
}

// a1Init は固有天賦1「実地観察ノート」を初期化する
// ルミがフィールドにいる間、すべての敵の岩元素耐性が15%減少する。
// Moonsign: Ascendant Gleam: さらに15%減少（合計30%）。
func (c *char) a1Init() {
	if c.Base.Ascension < 1 {
		return
	}
	c.a1Tick()
}

// a1Tick は60fごとに全敵へ岩元素耐性ダウンを適用する（hitlag非依存）
func (c *char) a1Tick() {
	if c.lumiActive {
		for _, t := range c.Core.Combat.Enemies() {
			e, ok := t.(*enemy.Enemy)
			if !ok {
				continue
			}
			// 基本の岩耐性-15%（90fで期限切れ → 60f毎のRefreshで常時維持）
			e.AddResistMod(combat.ResistMod{
				Base:  modifier.NewBase(a1GeoResKey, 90),
				Ele:   attributes.Geo,
				Value: -0.15,
			})
			// Ascendant Gleam: さらに-15%
			if c.MoonsignAscendant {
				e.AddResistMod(combat.ResistMod{
					Base:  modifier.NewBase(a1GeoResAscendKey, 90),
					Ele:   attributes.Geo,
					Value: -0.15,
				})
			}
		}
	}
	c.QueueCharTask(c.a1Tick, 60)
}

// a4Init は固有天賦4「万物博物図鑑」を初期化する
// アクティブキャラに応じてEMバフを付与する。
// リンネアの防御力の5%分のEMを増加。
// Moonsignキャラ → そのキャラのEM増加。
// 非Moonsignキャラ → リンネア自身のEM増加。
func (c *char) a4Init() {
	if c.Base.Ascension < 4 {
		return
	}

	// 各パーティメンバーに StatMod を付与し、Amount内でアクティブキャラを動的判定
	for _, char := range c.Core.Player.Chars() {
		idx := char.Index
		emMod := make([]float64, attributes.EndStatType)
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("linnea-a4-em", -1),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				// アクティブキャラでなければ適用しない
				if c.Core.Player.Active() != idx {
					return nil, false
				}
				// アクティブキャラがMoonsignかどうかを動的判定
				activeChar := c.Core.Player.ByIndex(idx)
				if !activeChar.StatusIsActive("moonsignKey") {
					return nil, false
				}
				emMod[attributes.EM] = c.TotalDef(false) * 0.05
				return emMod, true
			},
		})
	}

	// リンネア自身: 非Moonsignキャラがアクティブの場合にEM増加
	emModSelf := make([]float64, attributes.EndStatType)
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("linnea-a4-em-self", -1),
		AffectedStat: attributes.EM,
		Amount: func() ([]float64, bool) {
			activeChar := c.Core.Player.ByIndex(c.Core.Player.Active())
			if activeChar.StatusIsActive("moonsignKey") {
				return nil, false
			}
			emModSelf[attributes.EM] = c.TotalDef(false) * 0.05
			return emModSelf, true
		},
	})

	c.Core.Log.NewEvent("Linnea A4 EM sharing initialized", glog.LogCharacterEvent, c.Index)
}
