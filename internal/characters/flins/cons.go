package flins

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	c1EnergyICD    = "flins-c1-energy-icd"
	c2NorthlandKey = "flins-c2-northland"
	c2ResDebuff    = "flins-c2-res-debuff"
)

// C1
// The basic cooldown of the special Elemental Skill: Northland Spearstorm is reduced to 4s.
// Additionally, when party members trigger Lunar-Charged reactions, Flins will recover 8 Elemental Energy. This effect can occur once every 5.5s.
func (c *char) c1() {
	if c.Base.Cons < 1 {
		c.northlandCD = 6 * 60
		return
	}
	c.northlandCD = 4 * 60

	// Subscribe to Lunar-Charged reaction events for energy recovery
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.AttackTag != attacks.AttackTagLCDamage {
			return false
		}
		if c.StatusIsActive(c1EnergyICD) {
			return false
		}
		c.AddEnergy("flins-c1", 8)
		c.AddStatus(c1EnergyICD, 5.5*60, true)
		return false
	}, "flins-c1-energy")
}

// C2
// For the next 6s after using the special Elemental Skill: Northland Spearstorm, when Flins's next Normal Attack hits an opponent, it will deal an additional 50% of Flins's ATK as AoE Electro DMG. This DMG is considered Lunar-Charged DMG.
// When the moonsign is Moonsign: Ascendant Gleam, While Flins is on the field, after his Electro attacks hit an opponent, that opponent's Electro RES will be decreased by 25% for 7s.
func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}

	// Part 1: Additional damage after Northland Spearstorm
	// This is handled via a callback that checks for c2NorthlandKey status
	// The callback will be added to Normal Attacks in the attackE() function

	// Part 2: Electro RES decrease when Moonsign: Ascendant Gleam is active
	if !c.MoonsignAscendant {
		return
	}

	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != c.Index {
			return false
		}
		if atk.Info.Element != attributes.Electro {
			return false
		}
		if c.Core.Player.Active() != c.Index {
			return false
		}

		enemy, ok := args[0].(combat.Enemy)
		if !ok {
			return false
		}
		enemy.AddResistMod(combat.ResistMod{
			Base:  modifier.NewBaseWithHitlag(c2ResDebuff, 7*60),
			Ele:   attributes.Electro,
			Value: -0.25,
		})
		return false
	}, "flins-c2-res")
}

// C2 callback for additional damage after Northland Spearstorm
func (c *char) c2AdditionalDamage() combat.AttackCBFunc {
	if c.Base.Cons < 2 {
		return nil
	}

	done := false
	return func(a combat.AttackCB) {
		if done {
			return
		}
		if !c.StatusIsActive(c2NorthlandKey) {
			return
		}
		_, ok := a.Target.(combat.Enemy)
		if !ok {
			return
		}

		done = true
		c.DeleteStatus(c2NorthlandKey)

		// Additional 50% ATK as Electro DMG (Lunar-Charged)
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Flins C2 Dummy",
			FlatDmg:    0,
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 99), 0, 0)
	}
}

// C4
// Flins's ATK is increased by 20%.
// Additionally, his Ascension Talent "Whispering Flame" is changed: Flins's Elemental Mastery is increased by 10% of his ATK. The maximum increase obtainable this way is 220.
func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}

	// ATK 20% increase
	m := make([]float64, attributes.EndStatType)
	c.AddStatMod(character.StatMod{
		Base: modifier.NewBaseWithHitlag("Flins C4 ATK", -1),
		Amount: func() ([]float64, bool) {
			m[attributes.ATKP] = 0.20
			return m, true
		},
	})
}

// C6
// The DMG dealt to opponents by Flins's Lunar-Charged reactions is multiplied by 35%.
// When the moonsign is Moonsign: Ascendant Gleam, All nearby party members' Lunar-Charged DMG is multiplied by 10%.
func (c *char) c6() {
	if c.Base.Cons < 6 {
		return
	}

	// Flins's own Lunar-Charged DMG bonus: 35%
	c.AddElevationMod(character.ElevationMod{
		Base: modifier.NewBase("Flins C6", -1),
		Amount: func(ai combat.AttackInfo) (float64, bool) {
			if ai.AttackTag == attacks.AttackTagLCDamage {
				return 0.35, false
			} else {
				return 0, false
			}
		},
	})

	// Team-wide Lunar-Charged DMG bonus when Moonsign: Ascendant Gleam: 10%
	if c.MoonsignAscendant {
		for _, char := range c.Core.Player.Chars() {
			char.AddElevationMod(character.ElevationMod{
				Base: modifier.NewBase("Flins C6 Team", -1),
				Amount: func(ai combat.AttackInfo) (float64, bool) {
					if ai.AttackTag == attacks.AttackTagLCDamage {
						return 0.1, false
					} else {
						return 0, false
					}
				},
			})
		}
	}
}
