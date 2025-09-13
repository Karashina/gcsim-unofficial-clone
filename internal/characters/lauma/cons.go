package lauma

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

// C1: After using Elemental Skill, Normal Attack damage is increased by 40% for 8s
func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}
	
	c.Core.Events.Subscribe(event.OnSkill, func(args ...interface{}) bool {
		if c.Core.Player.Active() != c.Index {
			return false
		}
		
		c.AddAttackMod(character.AttackMod{
			Base: modifier.NewBaseWithHitlag("lauma-c1", 8*60),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				if atk.Info.AttackTag == attacks.AttackTagNormal {
					buff := make([]float64, attributes.EndStatType)
					buff[attributes.DmgP] = 0.4
					return buff, true
				}
				return nil, false
			},
		})
		
		return false
	}, "lauma-c1")
}

// C2: Elemental Skill grants additional charge
func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}
	// Implementation for additional skill charge would go here
}

// C4: Elemental Burst restores energy to team
func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}
	// Implementation for team energy restoration would go here
}

// C6: Enhanced burst damage and effects  
func (c *char) c6() {
	if c.Base.Cons < 6 {
		return
	}
	// Implementation for enhanced burst would go here
}