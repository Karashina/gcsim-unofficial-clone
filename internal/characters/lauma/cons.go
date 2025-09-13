package lauma

import (
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
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
	// Add a second charge to elemental skill
	c.SetNumCharges(action.ActionSkill, 2)
}

// C4: Elemental Burst restores energy to team
func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}
	
	c.Core.Events.Subscribe(event.OnBurst, func(args ...interface{}) bool {
		if c.Core.Player.Active() != c.Index {
			return false
		}
		
		// Restore 15 energy to all team members except self
		for i, char := range c.Core.Player.Chars() {
			if i != c.Index {
				char.AddEnergy("lauma-c4", 15)
			}
		}
		
		return false
	}, "lauma-c4")
}

// C6: Enhanced burst damage and effects  
func (c *char) c6() {
	if c.Base.Cons < 6 {
		return
	}
	
	// C6: Burst creates additional pyro damage over time
	c.Core.Events.Subscribe(event.OnBurst, func(args ...interface{}) bool {
		if c.Core.Player.Active() != c.Index {
			return false
		}
		
		// Create damage over time effect for 6 seconds
		for i := 1; i <= 6; i++ {
			c.Core.Tasks.Add(func() {
				ai := combat.AttackInfo{
					ActorIndex: c.Index,
					Abil:       "Elemental Burst (C6)",
					AttackTag:  attacks.AttackTagElementalBurst,
					ICDTag:     attacks.ICDTagNone,
					ICDGroup:   attacks.ICDGroupDefault,
					StrikeType: attacks.StrikeTypeDefault,
					Element:    attributes.Pyro,
					Durability: 25,
					Mult:       0.5, // 50% ATK per tick
				}
				
				c.Core.QueueAttack(
					ai,
					combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 2}, 3.0),
					0,
					0,
				)
			}, i*60) // Every second for 6 seconds
		}
		
		return false
	}, "lauma-c6")
}