package linnea

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
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
// ルミがフィールドにいる間、ルミ付近の敵の岩元素耐性が15%減少する。
// Moonsign: Ascendant Gleam: さらに15%減少（合計30%）。
func (c *char) a1Init() {
	if c.Base.Ascension < 1 {
		return
	}

	// ルミがアクティブな間、定期的に岩元素耐性ダウンを適用する
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		if !c.lumiActive {
			return false
		}
		ae, ok := args[1].(*combat.AttackEvent)
		if !ok {
			return false
		}
		// リンネアの攻撃のみ（ルミの攻撃 = リンネアのActorIndex）
		if ae.Info.ActorIndex != c.Index {
			return false
		}

		t, ok := args[0].(*enemy.Enemy)
		if !ok {
			return false
		}

		// 基本の岩耐性-15%
		t.AddResistMod(combat.ResistMod{
			Base:  modifier.NewBaseWithHitlag(a1GeoResKey, 3*60),
			Ele:   attributes.Geo,
			Value: -0.15,
		})

		// Ascendant Gleam: さらに-15%
		if c.MoonsignAscendant {
			t.AddResistMod(combat.ResistMod{
				Base:  modifier.NewBaseWithHitlag(a1GeoResAscendKey, 3*60),
				Ele:   attributes.Geo,
				Value: -0.15,
			})
		}
		return false
	}, "linnea-a1-geo-res")
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

	// 5% DEF による EM増加を各パーティメンバーに動的に適用
	for _, char := range c.Core.Player.Chars() {
		idx := char.Index
		isMoonsign := char.StatusIsActive("moonsignKey")

		if isMoonsign {
			// Moonsignキャラ: そのキャラのEM増加
			emMod := make([]float64, attributes.EndStatType)
			char.AddStatMod(character.StatMod{
				Base:         modifier.NewBase("linnea-a4-em", -1),
				AffectedStat: attributes.EM,
				Amount: func() ([]float64, bool) {
					// アクティブキャラの場合のみ適用
					if c.Core.Player.Active() != idx {
						return nil, false
					}
					emMod[attributes.EM] = c.TotalDef(false) * 0.05
					return emMod, true
				},
			})
		}
	}

	// 非Moonsignキャラがアクティブの場合、リンネア自身のEM増加
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
