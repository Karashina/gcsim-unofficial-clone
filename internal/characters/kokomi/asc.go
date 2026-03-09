package kokomi

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 固有天賦2 - +25%治癒効果、-100%会心率をステータスに恒久的に適用
func (c *char) passive() {
	m := make([]float64, attributes.EndStatType)
	m[attributes.Heal] = .25
	m[attributes.CR] = -1
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("kokomi-passive", -1),
		AffectedStat: attributes.NoStat,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})
}

// 海人の羽衣を使用した際、心海自身の「化海月」がフィールド上に存在する場合、持続時間がリフレッシュされる。
//
// - burst.goで突破レベルを確認しており、レベル不足時の無駄なキューイングを防止
func (c *char) a1() {
	if c.Core.Status.Duration("kokomiskill") <= 0 {
		return
	}
	// +1で同一フレームの期限切れ問題を回避
	c.Core.Status.Add("kokomiskill", 12*60+1)
}

// 海人の羽衣中、珊瑚宮心海のHP上限に基づく通常攻撃・重撃ダメージボーナスが
// 治癒効果の15%分さらに増加する。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != c.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagNormal && atk.Info.AttackTag != attacks.AttackTagExtra {
			return false
		}
		if c.Core.Status.Duration("kokomiburst") == 0 {
			return false
		}

		a4Bonus := c.Stat(attributes.Heal) * 0.15 * c.MaxHP()
		atk.Info.FlatDmg += a4Bonus

		return false
	}, "kokomi-a4")
}
