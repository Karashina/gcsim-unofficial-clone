package amber

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 2凸
// 手動起爆による爆発は追加で200%のダメージを与える。
func (c *char) c2() {
	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = 2
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("amber-c2", -1),
		Amount: func(atk *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagElementalArt {
				return nil, false
			}
			if atk.Info.Abil != manualExplosionAbil {
				return nil, false
			}
			return m, true
		},
	})
}
