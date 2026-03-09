package aloy

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// アーロイが凍られた野性でコイル効果を受けると、攻撃力が16%増加し、近くのパーティメンバーの攻撃力が8%増加する。
// この効果は10秒間持続する。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	for _, char := range c.Core.Player.Chars() {
		m := make([]float64, attributes.EndStatType)
		m[attributes.ATKP] = .08
		if char.Index == c.Index {
			m[attributes.ATKP] = .16
		}
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("aloy-a1", rushingIceDuration),
			AffectedStat: attributes.ATKP,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}
}

// アーロイが凍られた野性によるラッシュアイス状態の時、1秒ごとに氷元素ダメージバフが3.5%増加する。
// この方法で最大35%の氷元素ダメージバフ増加を獲得できる。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	stacks := 1
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag("aloy-strong-strike", rushingIceDuration),
		AffectedStat: attributes.CryoP,
		Amount: func() ([]float64, bool) {
			if stacks > 10 {
				stacks = 10
			}
			m[attributes.CryoP] = float64(stacks) * 0.035
			return m, true
		},
	})

	for i := 0; i < 10; i++ {
		// 1秒ごと、アーロイはヒットラグの影響を受けないのでこの方法で問題ない
		c.Core.Tasks.Add(func() { stacks++ }, 60*(1+i))
	}
}
