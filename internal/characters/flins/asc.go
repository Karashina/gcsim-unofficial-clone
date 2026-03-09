package flins

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// A0: 全キャラクターに基本反応ボーナスモディファイアを追加
func (c *char) a0() {
	for _, char := range c.Core.Player.Chars() {
		char.AddStatus("LC-Key", -1, false)
		char.AddLCBaseReactBonusMod(character.LCBaseReactBonusMod{
			Base: modifier.NewBase("Old World Secrets (A0)", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				maxval := 0.14
				return min(maxval, c.TotalAtk()/100*0.007), false
			},
		})
	}
}

// 固有天賦1
// ムーンサインが「ムーンサイン: 昇詼の輝き」の場合、Flinsが発動するルナチャージ反応のダメージが追加で20%増加する。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	if c.MoonsignAscendant {
		// ムーンサイン: 昇詼の輝き
		c.AddLCReactBonusMod(character.LCReactBonusMod{
			Base: modifier.NewBase("Symphony of Winter (A1)", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				return 0.2, false
			},
		})
	}
}

// 固有天賦2
// Flinsの元素熟知が攻撃力の8%分増加する。この方法で得られる最大増加量は160。
// 4命ノ星座で変更: Flinsの元素熟知が攻撃力の10%分増加する。この方法で得られる最大増加量は220。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag("Whispering Flame (A4)", -1),
		AffectedStat: attributes.EM,
		Extra:        true,
		Amount: func() ([]float64, bool) {
			m := make([]float64, attributes.EndStatType)
			if c.Base.Cons >= 4 {
				// 4命ノ星座: 攻撃力の10%, 最大220
				m[attributes.EM] = min(220, c.TotalAtk()*0.10)
			} else {
				// 基本固有天賢4: 攻撃力の8%, 最大160
				m[attributes.EM] = min(160, c.TotalAtk()*0.08)
			}
			return m, true
		},
	})
}
