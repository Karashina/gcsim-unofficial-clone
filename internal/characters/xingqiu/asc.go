package xingqiu

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 固有天賦1は未実装:
// TODO: 雨すだれの剣が破壊された時または持続時間が切れた時、行秋の最大HPの6%に基づき現在のキャラクターのHPを回復する。

// 行秋は水元素ダメージボーナスが20%増加する。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.HydroP] = 0.2
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("xingqiu-a4", -1),
		AffectedStat: attributes.HydroP,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})
}
