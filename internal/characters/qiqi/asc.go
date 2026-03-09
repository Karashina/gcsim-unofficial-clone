package qiqi

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 仙法·寒病鬼差の効果を受けているキャラクターが元素反応をトリガーしたとき、
// 受ける治療ボーナスが8秒間20%増加する。
// - イベントフックと受治療ボーナス関数を実装
// - TODO: AddIncHealBonusを開始時に追加し、七七の元素スキル使用時にイベント購読を行う方法に変更できる可能性あり
// - TODO: イベント購読を常時維持しない方が効率的だが、現状は明確さのためまとめている
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	a1Hook := func(args ...interface{}) bool {
		if c.StatusIsActive(skillBuffKey) {
			return false
		}
		atk := args[1].(*combat.AttackEvent)

		// アクティブキャラが七七の元素スキル効果を受けている唯一のキャラ
		active := c.Core.Player.ActiveChar()
		if atk.Info.ActorIndex != active.Index {
			return false
		}

		active.AddHealBonusMod(character.HealBonusMod{
			Base: modifier.NewBaseWithHitlag("qiqi-a1", 8*60),
			Amount: func() (float64, bool) {
				return .2, false
			},
		})

		return false
	}

	for i := event.ReactionEventStartDelim + 1; i < event.OnShatter; i++ {
		c.Core.Events.Subscribe(i, a1Hook, "qiqi-a1")
	}
}

// 固有天賦4はburst.goに実装:
// 七七が通常攻撃と重撃で敵に命中したとき、
// 50%の確率で敵に寿命の箓を付与する（6秒間）。
// この効果は30秒に1回のみ発動。
const a4ICDKey = "qiqi-a4-icd"
