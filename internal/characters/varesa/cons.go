package varesa

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/core/targets"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) c2CB() func(combat.AttackCB) {
	if c.Base.Cons < 2 {
		return nil
	}

	done := false
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if done {
			return
		}
		done = true
		c.AddEnergy("varesa-c2", 11.5)
	}
}

func (c *char) c4Plunge() {
	if c.Base.Cons < 4 {
		c.c4buff = 0
		return
	}
	if c.StatusIsActive(c4Key) {
		c.c4buff = min(20000, c.TotalAtk()*5)
		c.DeleteStatus(c4Key)
	} else {
		c.c4buff = 0
	}
}
func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = 1.0
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("varesa-c4", 45),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			return m, true
		},
	})
}

func (c *char) c6() {
	if c.Base.Cons < 6 {
		return
	}

	m := make([]float64, attributes.EndStatType)
	m[attributes.CR] = 0.1
	m[attributes.CD] = 1.0
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("varesa-c6", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			switch {
			case atk.Info.AttackTag == attacks.AttackTagElementalBurst:
			case atk.Info.AttackTag == attacks.AttackTagPlunge && atk.Info.Durability > 0: // TODO: collision?
			case atk.Info.Abil == kablamAbil:
			default:
				return nil, false
			}
			return m, true
		},
	})

}
