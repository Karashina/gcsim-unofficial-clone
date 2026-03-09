package xinyan

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/shield"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 情熱伝導が各レベルのシールドを発動するために必要な敵の数を減少させる。
//
// - シールドレベル2: 導入の必要数が敵1体に減少。
//
// - シールドレベル3: 熱狂の必要数が敵2体以上に減少。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	c.shieldLevel2Requirement -= 1
	c.shieldLevel3Requirement -= 1
}

// 情熱伝導のシールドで保護されたキャラクターの物理ダメージが15%増加する。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.PhyP] = 0.15
	for i, char := range c.Core.Player.Chars() {
		idx := i
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("xinyan-a4", -1),
			Amount: func(_ *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
				shd := c.Core.Player.Shields.Get(shield.XinyanSkill)
				if shd == nil {
					return nil, false
				}
				return m, c.Core.Player.Active() == idx
			},
		})
	}
}
