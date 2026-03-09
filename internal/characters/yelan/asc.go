package yelan

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// パーティの元素タイプが1/2/3/4種類の場合、夜蘭のHP上限が6%/12%/18%/30%上昇する。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	partyEleTypes := make(map[attributes.Element]bool)
	for _, char := range c.Core.Player.Chars() {
		partyEleTypes[char.Base.Element] = true
	}
	count := len(partyEleTypes)

	m := make([]float64, attributes.EndStatType)
	m[attributes.HPP] = float64(count) * 0.06
	if count >= 4 {
		m[attributes.HPP] = 0.3
	}

	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("yelan-a1", -1),
		AffectedStat: attributes.HPP,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})
}

// 玉瓒一擲が発動中、自身のフィールド上キャラクターのダメージが1%上昇する。
// 毎秒さらに3.5%ずつ上昇する。ダメージ上昇の最大値は50%。
// 持続時間中に深境カードを再発動すると、既存の効果は消去される。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	started := c.Core.F
	for _, char := range c.Core.Player.Chars() {
		this := char
		this.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("yelan-a4", 15*60),
			Amount: func(_ *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
				// キャラクターがフィールド上にいる必要がある
				if c.Core.Player.Active() != this.Index {
					return nil, false
				}
				// 経過時間を切り捨て
				dmg := float64((c.Core.F-started)/60)*0.035 + 0.01
				if dmg > 0.5 {
					dmg = 0.5
				}
				c.a4buff[attributes.DmgP] = dmg
				return c.a4buff, true
			},
		})
	}
}
