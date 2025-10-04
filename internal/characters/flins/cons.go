package flins

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/core/targets"
	"github.com/genshinsim/gcsim/pkg/modifier"
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
	c.Core.Events.Subscribe(event.OnLCReaction, func(args ...interface{}) bool {
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

		trg := args[0].(combat.Target)
		trg.AddResistMod(combat.ResistMod{
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
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}

		done = true
		c.DeleteStatus(c2NorthlandKey)

		// Additional 50% ATK as Electro DMG (Lunar-Charged)
		ai := combat.AttackInfo{
			ActorIndex:       c.Index,
			Abil:             "C2 Additional DMG",
			AttackTag:        attacks.AttackTagLCDamage,
			StrikeType:       attacks.StrikeTypeDefault,
			Element:          attributes.Electro,
			IgnoreDefPercent: 1,
		}
		em := c.Stat(attributes.EM)
		ai.FlatDmg = (c.TotalAtk() * 0.5 * (1 + c.LCBaseReactBonus(ai))) * (1 + ((6 * em) / (2000 + em)) + c.LCReactBonus(ai)) * 3

		snap := combat.Snapshot{
			CharLvl: c.Base.Level,
		}
		snap.Stats[attributes.CR] = c.Stat(attributes.CR)
		snap.Stats[attributes.CD] = c.Stat(attributes.CD)

		c.Core.QueueAttackWithSnap(
			ai,
			snap,
			combat.NewCircleHitOnTarget(a.Target, nil, 5),
			0,
		)
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
	c.AddLCReactBonusMod(character.LCReactBonusMod{
		Base: modifier.NewBase("Flins C6", -1),
		Amount: func(ai combat.AttackInfo) (float64, bool) {
			return 0.35, false
		},
	})

	// Team-wide Lunar-Charged DMG bonus when Moonsign: Ascendant Gleam: 10%
	if c.MoonsignAscendant {
		for _, char := range c.Core.Player.Chars() {
			char.AddLCReactBonusMod(character.LCReactBonusMod{
				Base: modifier.NewBase("Flins C6 Team", -1),
				Amount: func(ai combat.AttackInfo) (float64, bool) {
					return 0.10, false
				},
			})
		}
	}
}
