package dendro

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 固有天賦1の突破段階チェックはburst.go内で一度行われる
const a1Key = "dmc-a1"

// リーフロータスランプはフィールド上に存在する間、毎秒Overflowing Lotuslightを1レベル獲得する。
//
// - キャラ交代時に解除 - Kolibriより
func (c *Traveler) a1Init() {
	if c.Base.Ascension < 1 {
		return
	}
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(args ...interface{}) bool {
		prev := args[0].(int)
		prevChar := c.Core.Player.ByIndex(prev)
		prevChar.DeleteStatMod(a1Key)
		return false
	}, "dmc-a1-remove")
}

// 範囲内のアクティブキャラクターの元素熟知を6増加させる。
func (c *Traveler) a1Buff(delay int) {
	m := make([]float64, attributes.EndStatType)
	// 固有天賦1/6凸のバフは0.3秒ごとにティックし、1秒間適用。おそらくガジェット出現からカウント - Kolibriより
	c.Core.Tasks.Add(func() {
		if c.Core.Status.Duration(burstKey) <= 0 {
			return
		}
		if !c.Core.Combat.Player().IsWithinArea(combat.NewCircleHitOnTarget(c.burstPos, nil, c.burstRadius)) {
			return
		}
		m[attributes.EM] = float64(6 * c.burstOverflowingLotuslight)
		active := c.Core.Player.ActiveChar()
		active.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(a1Key, 60),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}, delay)
}

// Overflowing Lotuslightの最大スタック数は10。
func (c *Traveler) a1Stack(delay int) {
	c.Core.Tasks.Add(func() {
		if c.Core.Status.Duration(burstKey) > 0 && c.burstOverflowingLotuslight < 10 { // 元素爆発が未終了かつスタック上限未到達
			c.burstOverflowingLotuslight += 1
		}
	}, delay)
}

// 旅人の元素熟知1ポイントにつき、草薪の刃のダメージが0.15%、激流の結実のダメージが0.1%増加する。
func (c *Traveler) a4Init() {
	if c.Base.Ascension < 4 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("dmc-a4", -1),
		Amount: func(atk *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
			switch atk.Info.AttackTag {
			case attacks.AttackTagElementalArt:
				m[attributes.DmgP] = c.Stat(attributes.EM) * 0.0015
				return m, true
			case attacks.AttackTagElementalBurst:
				m[attributes.DmgP] = c.Stat(attributes.EM) * 0.001
				return m, true
			default:
				return nil, false
			}
		},
	})
}
