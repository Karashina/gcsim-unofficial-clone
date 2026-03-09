package mika

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 星霜の旋の霜流れの矢が初めて敵に命中した時、または霜星弾が敵に命中した時、
// 固有天賦「索敵弾幕」のディテクタースタックが1つ生成される。
// 固有天賦「索敵弾幕」のアンロックが必要。
func (c *char) c2() func(combat.AttackCB) {
	if c.Base.Cons < 2 || c.Base.Ascension < 1 {
		return nil
	}

	done := false
	return func(_ combat.AttackCB) {
		if done {
			return
		}

		c.addDetectorStack()
		done = true
	}
}

// 星霜の旋の極寒の風が獲得できるディテクタースタックの最大数が1つ増加する。
// 固有天賦「索敵弾幕」のアンロックが必要。
// さらに、極寒の風の効果を受けたアクティブキャラクターの物理会心ダメージが60%増加する。
func (c *char) c6(char *character.CharWrapper) {
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBaseWithHitlag("mika-c6", skillBuffDuration),
		Amount: func(atk *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
			if c.Core.Player.Active() != char.Index {
				return nil, false
			}
			if atk.Info.Element != attributes.Physical {
				return nil, false
			}
			return c.c6buff, true
		},
	})
}
