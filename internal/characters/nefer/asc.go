package nefer

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// A0
// Neferの元素熟知1ポイントにつき、Lunar-Bloomの基礎ダメージが0.0175%増加する（最大14%）。
func (c *char) a0() {
	for _, char := range c.Core.Player.Chars() {
		char.AddStatus("LB-Key", -1, false)
		char.AddLBBaseReactBonusMod(character.LBBaseReactBonusMod{
			Base: modifier.NewBase("Moonsign Benediction: Dusklit Eaves (A0)", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				maxval := 0.14
				return min(maxval, c.Stat(attributes.EM)*0.000175), false
			},
		})
	}
}

// A1：種/ヴェールメカニクス（簡略化）：種を吸収すると偽りのヴェールスタックを獲得し、閾値で元素熟知ボーナスとPPダメージがスタックあたり8%増加する。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	// 現在は重撃/幻影攻撃での直接種吸収で処理
}
