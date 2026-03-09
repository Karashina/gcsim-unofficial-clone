package venti

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	a0SwirlKey    = "venti-a0-swirl"
	a0BurstDmgKey = "venti-a0-burst-dmg"
)

// a0HexereiInitは以下を登録する：
//  1. Ventiに永続AttackModを付与し、a0拡散バフがアクティブ中に元素爆発攻撃に+35% DmgPを与える。
//  2. 拡散イベントサブスクリプション：元素爆発の眼がアクティブ中に任意のキャラが
//     拡散を発動すると、そのキャラは4秒間+50% DmgPを得て、Ventiはa0拡散バフを4秒間得る。
func (c *char) a0HexereiInit() {
	// A0はHexereiモードとパーティに2人以上のHexereiキャラが必要
	if !c.isHexerei || !c.hasHexBonus {
		return
	}
	// 拡散バフがアクティブ中の永続元素爆発攻撃ボーナス（+35%）
	m := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase(a0BurstDmgKey, -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.ActorIndex != c.Index {
				return nil, false
			}
			if atk.Info.AttackTag != attacks.AttackTagElementalBurst {
				return nil, false
			}
			if !c.StatusIsActive(a0SwirlKey) {
				return nil, false
			}
			for i := range m {
				m[i] = 0
			}
			m[attributes.DmgP] = 0.35
			return m, true
		},
	})

	// 4つの拡散イベントをサブスクライブ
	swirlEvents := []event.Event{
		event.OnSwirlHydro,
		event.OnSwirlPyro,
		event.OnSwirlCryo,
		event.OnSwirlElectro,
	}
	for _, ev := range swirlEvents {
		ev := ev // capture
		c.Core.Events.Subscribe(ev, func(args ...interface{}) bool {
			// Ventiの元素爆発の眼がアクティブな時のみ発動
			if c.Core.F >= c.burstEnd {
				return false
			}
			atk, ok := args[1].(*combat.AttackEvent)
			if !ok {
				return false
			}
			// トリガーしたキャラに4秒間（240f）+50% DmgP AttackModを適用
			chars := c.Core.Player.Chars()
			actorIdx := atk.Info.ActorIndex
			if actorIdx >= 0 && actorIdx < len(chars) {
				triggerBuff := make([]float64, attributes.EndStatType)
				triggerBuff[attributes.DmgP] = 0.50
				chars[actorIdx].AddAttackMod(character.AttackMod{
					Base: modifier.NewBase(
						fmt.Sprintf("venti-a0-swirl-dmg-%d", actorIdx), 240,
					),
					Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
						return triggerBuff, true
					},
				})
			}
			// Ventiのa0拡散ステータスを4秒間有効化
			c.AddStatus(a0SwirlKey, 240, false)
			return false
		}, fmt.Sprintf("venti-a0-%v", ev))
	}
}

// A1は実装されておらず、今後も実装される可能性は低い：
// 長押し「高天の歌」は20秒間持続する上昇気流を生成する。

// 「風神の詩」の効果終了後、Ventiの元素エネルギーを15回復する。
// 元素変化が発生した場合、その元素に対応するパーティ全員の元素エネルギーもそれぞれ15回復する。
//
// - 突破レベルチェックはburst.goで行い、失敗時にキューに入れるのを避ける
func (c *char) a4() {
	c.AddEnergy("venti-a4", 15)
	if c.qAbsorb == attributes.NoElement {
		return
	}
	for _, char := range c.Core.Player.Chars() {
		if char.Base.Element == c.qAbsorb {
			char.AddEnergy("venti-a4", 15)
		}
	}
}
