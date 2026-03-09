package sayu

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 2凸:
// よぶぶきの術・風隠急進が以下の効果を得る:
// 単押しモードのれんこつむじかぜキックのダメージが3.3%上昇する。
// ふうふう風車状態の0.5秒ごとに、れんこつむじかぜキックの
// ダメージが3.3%上昇する。この方法での最大ダメージ上昇は66%。
func (c *char) c2() {
	m := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("sayu-c2", -1),
		Amount: func(atk *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
			if atk.Info.ActorIndex != c.Index {
				return nil, false
			}
			if atk.Info.AttackTag != attacks.AttackTagElementalArt {
				return nil, false
			}
			m[attributes.DmgP] = c.c2Bonus
			return m, true
		},
	})
}
