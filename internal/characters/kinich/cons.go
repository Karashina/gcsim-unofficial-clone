package kinich

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	c4ICDKey = "kinich-c4-icd"
)

func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.CD] = 1.0
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("kinich-c1", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.Abil != "Canopy Hunter: Riding High (Scalespiker Cannon DMG)" {
				return nil, false
			}
			return m, true
		},
	})
}

func (c *char) c2() {
	if c.Base.Cons < 2 {
		c.sscradius = 2.5
		return
	}
	c.sscradius = 5
	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = 1.0
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("kinich-c2", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.Abil != "Auspicious Beast's Shape (Scalespiker Cannon Bounce)" && atk.Info.Abil != "Canopy Hunter: Riding High (Scalespiker Cannon DMG)" || !c.StatusIsActive(c2buffKey) {
				return nil, false
			}
			return m, true
		},
	})
}

func (c *char) c2CB() combat.AttackCBFunc {
	if c.Base.Cons < 2 {
		return nil
	}
	return func(a combat.AttackCB) {
		if a.Damage == 0 {
			return
		}
		e, ok := a.Target.(*enemy.Enemy)
		if !ok {
			return
		}
		e.AddResistMod(combat.ResistMod{
			Base:  modifier.NewBaseWithHitlag("kinich-c2-res", 6*60),
			Ele:   attributes.Dendro,
			Value: -0.3,
		})
	}
}

func (c *char) c4energy() {
	if c.Base.Cons < 4 {
		return
	}
	if c.StatusIsActive(c4ICDKey) {
		return
	}
	c.AddEnergy("kinich-c4", 5)
	c.AddStatus(c4ICDKey, 2.8*60, true)
}

func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = 0.7
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBaseWithHitlag("kinich-c4-burstbuff", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			switch atk.Info.AttackTag {
			case attacks.AttackTagElementalBurst:
				return m, true
			default:
				return nil, false
			}
		},
	})
}
