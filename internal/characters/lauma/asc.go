package lauma

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

// A1: Basic passive - increases pyro damage when HP is above 50%
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	
	// Add a simple passive that increases pyro damage based on current HP
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("lauma-a1", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if c.CurrentHPRatio() > 0.5 && atk.Info.Element == attributes.Pyro {
				buff := make([]float64, attributes.EndStatType)
				buff[attributes.DmgP] = 0.15 // 15% damage increase
				return buff, true
			}
			return nil, false
		},
	})
}

// A4: Advanced passive - provides team pyro damage bonus
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	
	// Team-wide pyro damage bonus when Lauma uses elemental skill
	// This would be implemented when skill mechanics are expanded
}