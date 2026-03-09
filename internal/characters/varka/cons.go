package varka

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// C1 - "Come, Friend, Let Us Dance Beneath the Moon's Soft Glow"
// S&D突入時にFWAチャージを1つ即座に付与し、Lyrical Libation効果を付与
// enterSturmUndDrangで処理（即時チャージ付与 + c1LyricalKeyステータス）
// 200%ダメージ倍率はskill.goのfourWindsAscensionとcharge.goのazureDevourで処理

// C2 - "When Dawn Breaks, Our Journey Shall Take Flight"
// FWAまたはAzure Devour時にATKの800%に等しい追加風元素攻撃
// skill.goのc2Strikeとc2Initで処理

// C4 - "For None May Take From Us Our Freedom of Song"
// ヴァルカが拡散反応を発動すると、周囲のパーティメンバー全員が
// 10秒間、風元素ダメージボーナス20%と対応する元素ダメージボーナスを獲得する。
func (c *char) c4Init() {
	swirlMap := map[event.Event]attributes.Stat{
		event.OnSwirlHydro:   attributes.HydroP,
		event.OnSwirlPyro:    attributes.PyroP,
		event.OnSwirlCryo:    attributes.CryoP,
		event.OnSwirlElectro: attributes.ElectroP,
	}

	for ev, eleStat := range swirlMap {
		eleStat := eleStat // capture for closure
		c.Core.Events.Subscribe(ev, func(args ...interface{}) bool {
			atk := args[1].(*combat.AttackEvent)
			// ヴァルカが拡散を発動した時のみトリガー
			if atk.Info.ActorIndex != c.Index {
				return false
			}

			// 全パーティメンバーに適用
			for _, char := range c.Core.Player.Chars() {
				m := make([]float64, attributes.EndStatType)
				m[attributes.AnemoP] = 0.20
				m[eleStat] = 0.20

				char.AddStatMod(character.StatMod{
					Base:         modifier.NewBaseWithHitlag(c4Key, 10*60),
					AffectedStat: attributes.NoStat,
					Amount: func() ([]float64, bool) {
						return m, true
					},
				})
			}

			return false
		}, fmt.Sprintf("varka-c4-%v", ev))
	}
}

// C6 - "Beloved Mondstadt, Steadfast You Shall Shine"
// FWA→Azureチェーン、Azure→FWAチェーンをチャージ消費なしで実行
// ウィンドウステータス（c6FWAWindowKey, c6AzureWindowKey）は以下で設定:
//   - skill.go fourWindsAscension() → c6FWAWindowKeyを設定
//   - charge.go azureDevour() → c6AzureWindowKeyを設定
// fourWindsAscensionとazureDevourでそれぞれ消費される
//
// C6のA4スタックからの会心ダメージはasc.go a4Apply()で処理
