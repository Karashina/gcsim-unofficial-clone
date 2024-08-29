package mualani

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}
	c.WaveMomentum += 2
}

func (c *char) c2Puffer() {
	if c.Base.Cons < 2 {
		return
	}
	c.WaveMomentum++
	if c.pufferCount == 2 {
		c.AddNightsoul("mualani-c2", 12)
	}
}

func (c *char) c4energy() {
	if c.Base.Cons < 4 {
		return
	}
	c.AddEnergy("mualani-c4", 8)
}

func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBaseWithHitlag("mualani-c4-burstbuff", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			m[attributes.DmgP] = 0.75
			switch atk.Info.AttackTag {
			case attacks.AttackTagElementalBurst:
				return m, true
			default:
				return nil, false
			}
		},
	})
}
