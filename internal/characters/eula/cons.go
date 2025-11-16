package eula

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func (c *char) c4() {
	if c.Core.Combat.DamageMode {
		m := make([]float64, attributes.EndStatType)
		m[attributes.DmgP] = 0.25
		c.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("eula-c4", -1),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				if atk.Info.Abil != "Glacial Illumination (Lightfall)" {
					return nil, false
				}
				if !c.Core.Combat.DamageMode {
					return nil, false
				}
				x, ok := t.(*enemy.Enemy)
				if !ok {
					return nil, false
				}
				if x.HP()/x.MaxHP() >= 0.5 {
					return nil, false
				}
				return m, true
			},
		})
	}
}

