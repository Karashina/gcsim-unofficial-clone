package aubadeofmorningstarandmoon

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterSetFunc(keys.AubadeOfMorningstarAndMoon, NewSet)
}

type Set struct {
	Index int
	Count int
}

func (s *Set) SetIndex(idx int) { s.Index = idx }
func (s *Set) GetCount() int    { return s.Count }
func (s *Set) Init() error      { return nil }

// 曙星と月のオバード
// 2セット: 元素熟知+80
// 4セット: 装備キャラがフィールド外の場合、Lunar-Charged、Lunar-Bloom、
// およびLunar-Crystallizeボーナスが20%増加。
// パーティのMoonsignレベルがAscendant Gleam以上の場合、Lunar反応ダメージが
// さらに40%増加。
// この効果は装備キャラがフィールド上で3秒後に消失。

func NewSet(c *core.Core, char *character.CharWrapper, count int, param map[string]int) (info.Set, error) {
	s := Set{Count: count}

	if count >= 2 {
		// 2セット: 元素熟知+80
		m := make([]float64, attributes.EndStatType)
		m[attributes.EM] = 80
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("aubade-2pc", -1),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}

	if count >= 4 {
		const activeKey = "aubade-4pc-active-timer"
		activeTimeout := 180 // 3s

		// キャラクターがアクティブになったタイミングを追跡
		c.Events.Subscribe(event.OnCharacterSwap, func(args ...interface{}) bool {
			if c.Player.Active() != char.Index {
				return false
			}
			// キャラクターがアクティブになったらタイマーを開始
			char.AddStatus(activeKey, activeTimeout, true)
			return false
		}, "aubade-4pc-swap")

		// ルナリアクションダメージボーナス用の攻撃モディファイアを追加
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("aubade-4pc", -1),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				// Lunar反応ダメージかどうか確認
				if atk.Info.AttackTag != attacks.AttackTagLCDamage &&
					atk.Info.AttackTag != attacks.AttackTagLBDamage &&
					atk.Info.AttackTag != attacks.AttackTagLCrsDamage {
					return nil, false
				}

				// キャラがフィールド外かどうか確認（アクティブタイマーが無効）
				if char.StatusIsActive(activeKey) {
					return nil, false
				}

				val := make([]float64, attributes.EndStatType)

				// 基本ボーナス: フィールド外で20%
				bonus := 0.20

				// 追加ボーナス: MoonsignレベルがAscendant Gleam以上で40%
				// キャラにMoonsign Ascendantステータスがあるか確認
				if char.MoonsignAscendant {
					bonus += 0.40
				}

				val[attributes.DmgP] = bonus
				return val, true
			},
		})
	}

	return &s, nil
}
