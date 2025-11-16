package nefer

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// C4: Const - Verdant gain +25% on-field; while in Shadow Dance, nearby enemies have Dendro RES -20% (removed 4.5s after exit)
func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}
	// add onfield constraint
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(args ...interface{}) bool {
		prev := args[0].(int)
		next := args[1].(int)

		if prev == c.Index && c.c4buffkey {
			c.Core.Player.Verdant.SetGainBonus(c.Core.Player.Verdant.GetGainBonus() - 0.25)
			c.c4buffkey = false
		}
		if next == c.Index {
			c.Core.Player.Verdant.SetGainBonus(c.Core.Player.Verdant.GetGainBonus() + 0.25)
			c.c4buffkey = true
		}
		return false
	}, "nefer-c4-verdantdewgain")

	// Apply Dendro RES -20% to nearby opponents while in Shadow Dance.
	// When Nefer exits Shadow Dance, this effect lasts an additional 4.5s.
	// We detect skill use while Nefer is active and apply the resist mod to nearby enemies.
	c.Core.Events.Subscribe(event.OnSkill, func(args ...interface{}) bool {
		// only trigger if Nefer is the active character
		if c.Core.Player.Active() != c.Index {
			return false
		}

		// radius: use 10 units as "nearby" (common convention in other skills)
		enemies := c.Core.Combat.EnemiesWithinArea(combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 10), nil)
		if len(enemies) == 0 {
			return false
		}

		// duration: Shadow Dance base 10s (10*60) + 4.5s (270 frames) = 870 frames
		dur := 10*60 + 270
		for _, e := range enemies {
			targ, ok := e.(*enemy.Enemy)
			if !ok {
				continue
			}
			targ.AddResistMod(combat.ResistMod{
				Base:  modifier.NewBaseWithHitlag("nefer-c4-dendro", dur),
				Ele:   attributes.Dendro,
				Value: -0.20,
			})
		}

		return false
	}, "nefer-c4")
}

// C6: Converts 2nd PP hit and adds extra AoE LB DMG (EM-scaling). If Moonsign Ascendant, LB DMG +15% via ElevationMod.
func (c *char) c6() {
	if c.Base.Cons < 6 {
		return
	}

	// Moonsign: Ascendant Gleam - Lunar-Bloom DMG elevation by 15%
	// This is handled via ElevationBonus in character system
	if c.MoonsignAscendant {
		c.AddElevationMod(character.ElevationMod{
			Base: modifier.NewBase("Nefer C6", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				if ai.AttackTag == attacks.AttackTagLBDamage {
					return 0.15, false
				}
				return 0, false
			},
		})
	}
}

