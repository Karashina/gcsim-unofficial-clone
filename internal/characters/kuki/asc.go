package kuki

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 忍のHPが50%以下の場合、与える治癒効果+15%。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.Heal] = .15
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("kuki-a1", -1),
		AffectedStat: attributes.Heal,
		Amount: func() ([]float64, bool) {
			if c.CurrentHPRatio() <= 0.5 {
				return m, true
			}
			return nil, false
		},
	})
}

// 草薙の稲光の制約の輪のアビリティは忍の元素熟知に基づいて強化される:
//
// - 回復量が元素熟知の75%分増加。
func (c *char) a4Healing() float64 {
	if c.Base.Ascension < 4 {
		return 0
	}
	return c.Stat(attributes.EM) * 0.75
}

// 草薙の稲光の制約の輪のアビリティは忍の元素熟知に基づいて強化される:
//
// - ダメージ量が元素熟知の25%分増加。
func (c *char) a4Damage() float64 {
	if c.Base.Ascension < 4 {
		return 0
	}
	return c.Stat(attributes.EM) * 0.25
}
