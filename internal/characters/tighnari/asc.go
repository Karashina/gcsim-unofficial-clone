package tighnari

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// Tighnariが花筒矢を放った後、4秒間元素熔通が50上昇する。
//
// - aimed.goで聴天等級チェックを行い、失敗を回避するためここではチェックしない
func (c *char) a1() {
	m := make([]float64, attributes.EndStatType)
	m[attributes.EM] = 50
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("tighnari-a1", 4*60),
		AffectedStat: attributes.EM,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})
}

// Tighnariの元素熔通1ポイントにつき、重撃と織用素翠の祠矢のダメージが0.06%上昇する。
// この方法で得られるダメージボーナスの上限は60%。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("tighnari-a4", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagExtra && atk.Info.AttackTag != attacks.AttackTagElementalBurst {
				return nil, false
			}

			bonus := c.Stat(attributes.EM) * 0.0006
			if bonus > 0.6 {
				bonus = 0.6
			}
			m[attributes.DmgP] = bonus
			return m, true
		},
	})
}
