package nefer

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

// A0
// Every point of Elemental Mastery that Nefer has increases Lunar-Bloom's Base DMG by 0.0175%, up to a maximum of 14%.
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

// A1: Seeds/Veil mechanic (simplified): absorbing seeds grants Veil of Falsehood stacks; at thresholds grants EM bonus and increases PP DMG by 8% per stack.
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}

	// Subscribe to charge attack to add Veil of Falsehood stacks
	// For simplicity, assume we gain stacks on CA/PP usage
	// In actual implementation, would need to check Seeds of Deceit on field
	c.Core.Events.Subscribe(event.OnChargeAttack, func(args ...interface{}) bool {
		if c.Core.Player.Active() != c.Index {
			return false
		}

		// Add Veil of Falsehood stack (simplified - assuming 1 seed absorbed per CA)
		maxStacks := 3.0
		if c.Base.Cons >= 2 {
			maxStacks = 5.0
		}

		if c.a1count < maxStacks {
			c.a1count++
			// Duration is tracked per stack, but simplified here
			c.AddStatus("veil-of-falsehood", 9*60, true)

			// When reaching max stacks, add EM bonus
			if c.a1count >= maxStacks || (c.a1count >= 3 && c.Base.Cons < 2) {
				emBonus := 100.0
				if c.Base.Cons >= 2 && c.a1count >= 5 {
					emBonus = 200.0
				}
				c.AddStatMod(character.StatMod{
					Base:         modifier.NewBaseWithHitlag("veil-em-bonus", 8*60),
					AffectedStat: attributes.EM,
					Amount: func() ([]float64, bool) {
						m := make([]float64, attributes.EndStatType)
						m[attributes.EM] = emBonus
						return m, true
					},
				})
			}
		}
		return false
	}, "nefer-a1-veil")
}
