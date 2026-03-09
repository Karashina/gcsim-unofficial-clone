package yoimiya

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const a1Key = "yoimiya-a1"

// 玅扇火舞中、嬵宮の通常攻撃が命中すると、
// 炎元素ダメージバフが2%上昇する。この効果は3秒間持続し、
// 最大10スタックまで蓄積可能。
func (c *char) makeA1CB() combat.AttackCBFunc {
	if c.Base.Ascension < 1 {
		return nil
	}
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if c.Core.Player.Active() != c.Index {
			return
		}
		if !c.StatusIsActive(skillKey) {
			return
		}

		if !c.StatusIsActive(a1Key) {
			c.a1Stacks = 0
		}
		if c.a1Stacks < 10 {
			c.a1Stacks++
		}

		m := make([]float64, attributes.EndStatType)
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBase(a1Key, 3*60),
			AffectedStat: attributes.PyroP,
			Amount: func() ([]float64, bool) {
				m[attributes.PyroP] = float64(c.a1Stacks) * 0.02
				return m, true
			},
		})
	}
}

// 琉金の雲間草を使用すると、周囲のチームメンバー（嬵宮を除く）の
// 攻撃力が15秒間10%上昇する。さらに、琉金の雲間草使用時の
// 「花火丁子童」のスタック数に応じて攻撃力バフが追加される。
// 各スタックはこの攻撃力バフを1%上昇させる。
func (c *char) a4() {
	c.a4Bonus[attributes.ATKP] = 0.1 + float64(c.a1Stacks)*0.01
	for _, x := range c.Core.Player.Chars() {
		if x.Index == c.Index {
			continue
		}
		x.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("yoimiya-a4", 900),
			AffectedStat: attributes.ATKP,
			Amount: func() ([]float64, bool) {
				return c.a4Bonus, true
			},
		})
	}
}
