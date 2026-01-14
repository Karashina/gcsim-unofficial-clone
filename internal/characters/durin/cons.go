package durin

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	// C1 state keys
	c1CycleKey = "durin-c1-cycle"

	// C1 durations and values
	c1CycleDuration          = 20 * 60 // 20 seconds
	c1MaxStacks              = 20
	c1PurityFlatDmgPercent   = 0.60 // 60% of Durin's ATK
	c1DarknessFlatDmgPercent = 1.50 // 150% of Durin's ATK
	c1DarknessStackCost      = 2    // Consumes 2 stacks per burst hit
	c1NoConsumeChance        = 0.30 // C4: 30% chance to not consume stacks

	// C2 state keys
	c2BuffKey    = "durin-c2-buff"
	c2EleBuffKey = "durin-c2-ele-buff"

	// C2 durations and values
	c2BuffDuration = 20 * 60 // 20 seconds
	c2EleDuration  = 6 * 60  // 6 seconds
	c2DmgBonus     = 0.50    // 50% DMG bonus

	// C4 values
	c4BurstDmgBonus = 0.40 // 40% Elemental Burst DMG bonus

	// C6 state keys
	c6DefShredKey = "durin-c6-def-shred"

	// C6 durations and values
	c6DefShredDuration = 6 * 60 // 6 seconds
	c6DefShred         = 0.30   // 30% DEF shred
)

// C1: Adamah's Redemption
// After casting Elemental Burst:
// Purity: Other party members gain 20 stacks of Cycle of Enlightenment (20s)
//
//	When dealing NA/CA/Plunge/Skill/Burst DMG, consume 1 stack to add 60% of Durin's ATK as flat DMG
//
// Darkness: Durin gains 20 stacks of Cycle of Enlightenment (20s)
//
//	When dealing Burst DMG, consume 2 stacks to add 150% of Durin's ATK as flat DMG
func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}

	// Subscribe to OnEnemyHit for flat damage application
	c.Core.Events.Subscribe(event.OnEnemyHit, c.c1OnEnemyHit, "durin-c1-flat-dmg")
}

func (c *char) c1OnBurstPurity() {
	if c.Base.Cons < 1 {
		return
	}

	// Grant stacks to all other party members
	for _, char := range c.Core.Player.Chars() {
		if char.Index == c.Index {
			continue // Skip Durin
		}
		c.cycleStacks[char.Index] = c1MaxStacks
		c.cycleExpiry[char.Index] = c.Core.F + c1CycleDuration
	}

	c.Core.Log.NewEvent("C1: Cycle of Enlightenment granted to party (Purity)", glog.LogCharacterEvent, c.Index).
		Write("stacks", c1MaxStacks)
}

func (c *char) c1OnBurstDarkness() {
	if c.Base.Cons < 1 {
		return
	}

	// Grant stacks to Durin only
	c.cycleStacks[c.Index] = c1MaxStacks
	c.cycleExpiry[c.Index] = c.Core.F + c1CycleDuration

	c.Core.Log.NewEvent("C1: Cycle of Enlightenment granted to Durin (Darkness)", glog.LogCharacterEvent, c.Index).
		Write("stacks", c1MaxStacks)
}

func (c *char) c1OnEnemyHit(args ...interface{}) bool {
	if c.Base.Cons < 1 {
		return false
	}

	atk := args[1].(*combat.AttackEvent)
	charIndex := atk.Info.ActorIndex

	// Check if this character has stacks
	expiry, ok := c.cycleExpiry[charIndex]
	if !ok || c.Core.F >= expiry || c.cycleStacks[charIndex] <= 0 {
		return false
	}

	// Check if this is Durin's Darkness mode stacks (only applies to Burst DMG)
	if charIndex == c.Index {
		// Durin's stacks only work with Burst DMG (Darkness mode)
		if atk.Info.AttackTag != attacks.AttackTagElementalBurst {
			return false
		}

		// Consume 2 stacks
		stacksToConsume := c1DarknessStackCost

		// C4: 30% chance to not consume stacks
		if c.Base.Cons >= 4 && c.Core.Rand.Float64() < c1NoConsumeChance {
			stacksToConsume = 0
			c.Core.Log.NewEvent("C4: Cycle of Enlightenment stacks not consumed", glog.LogCharacterEvent, c.Index)
		}

		if c.cycleStacks[charIndex] < stacksToConsume {
			stacksToConsume = c.cycleStacks[charIndex]
		}
		c.cycleStacks[charIndex] -= stacksToConsume

		// Add flat damage: 150% of Durin's ATK
		flatDmg := c.TotalAtk() * c1DarknessFlatDmgPercent
		atk.Info.FlatDmg += flatDmg

		c.Core.Log.NewEvent("C1: Cycle of Enlightenment (Darkness) flat DMG added", glog.LogCharacterEvent, c.Index).
			Write("flat_dmg", flatDmg).
			Write("stacks_remaining", c.cycleStacks[charIndex])

		return false
	}

	// Other party members' stacks (Purity mode)
	// Check if the character is the active character
	if charIndex != c.Core.Player.Active() {
		return false
	}

	// Check if valid attack type (NA, CA, Plunge, Skill, Burst)
	validTag := atk.Info.AttackTag == attacks.AttackTagNormal ||
		atk.Info.AttackTag == attacks.AttackTagExtra ||
		atk.Info.AttackTag == attacks.AttackTagPlunge ||
		atk.Info.AttackTag == attacks.AttackTagElementalArt ||
		atk.Info.AttackTag == attacks.AttackTagElementalArtHold ||
		atk.Info.AttackTag == attacks.AttackTagElementalBurst

	if !validTag {
		return false
	}

	// Consume 1 stack
	stacksToConsume := 1

	// C4: 30% chance to not consume stacks
	if c.Base.Cons >= 4 && c.Core.Rand.Float64() < c1NoConsumeChance {
		stacksToConsume = 0
		c.Core.Log.NewEvent("C4: Cycle of Enlightenment stacks not consumed", glog.LogCharacterEvent, charIndex)
	}

	c.cycleStacks[charIndex] -= stacksToConsume

	// Add flat damage: 60% of Durin's ATK
	flatDmg := c.TotalAtk() * c1PurityFlatDmgPercent
	atk.Info.FlatDmg += flatDmg

	c.Core.Log.NewEvent("C1: Cycle of Enlightenment (Purity) flat DMG added", glog.LogCharacterEvent, charIndex).
		Write("flat_dmg", flatDmg).
		Write("stacks_remaining", c.cycleStacks[charIndex])

	return false
}

// C2: Unsound Visions
// For 20s after Durin uses Elemental Burst, after party members trigger
// Vaporize, Melt, Burning, Overloaded, Pyro Swirl, or Pyro Crystallize,
// or deal Pyro/Dendro DMG to burning opponents, all party members gain
// 50% Pyro DMG and corresponding elemental DMG bonus for 6s
func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}

	// Subscribe to reaction events
	c.Core.Events.Subscribe(event.OnVaporize, c.c2ReactionCB(attributes.Hydro), "durin-c2-vaporize")
	c.Core.Events.Subscribe(event.OnMelt, c.c2ReactionCB(attributes.Cryo), "durin-c2-melt")
	c.Core.Events.Subscribe(event.OnBurning, c.c2ReactionCB(attributes.Dendro), "durin-c2-burning")
	c.Core.Events.Subscribe(event.OnOverload, c.c2ReactionCB(attributes.Electro), "durin-c2-overload")
	c.Core.Events.Subscribe(event.OnSwirlPyro, c.c2ReactionCB(attributes.Anemo), "durin-c2-pyro-swirl")
	c.Core.Events.Subscribe(event.OnCrystallizePyro, c.c2ReactionCB(attributes.Geo), "durin-c2-pyro-crystallize")

	// Subscribe for Pyro/Dendro DMG to burning enemies
	c.Core.Events.Subscribe(event.OnEnemyDamage, c.c2OnDamageCB, "durin-c2-burning-dmg")
}

func (c *char) c2ReactionCB(otherEle attributes.Element) func(args ...interface{}) bool {
	return func(args ...interface{}) bool {
		// Check if within 20s of burst
		if !c.StatusIsActive(c2BuffKey) {
			return false
		}

		c.c2ApplyBuff(otherEle)
		return false
	}
}

func (c *char) c2OnDamageCB(args ...interface{}) bool {
	if !c.StatusIsActive(c2BuffKey) {
		return false
	}

	atk := args[1].(*combat.AttackEvent)
	target := args[0].(combat.Target)
	e, ok := target.(*enemy.Enemy)
	if !ok {
		return false
	}

	// Check if target is burning
	if !e.AuraContains(attributes.Pyro, attributes.Dendro) {
		return false
	}

	// Check if dealing Pyro or Dendro DMG
	if atk.Info.Element != attributes.Pyro && atk.Info.Element != attributes.Dendro {
		return false
	}

	c.c2ApplyBuff(atk.Info.Element)
	return false
}

func (c *char) c2ApplyBuff(otherEle attributes.Element) {
	// Apply 50% Pyro DMG bonus and corresponding elemental DMG bonus to all party members
	for _, char := range c.Core.Player.Chars() {
		// Pyro DMG bonus
		m := make([]float64, attributes.EndStatType)
		m[attributes.PyroP] = c2DmgBonus
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(c2EleBuffKey+"-pyro", c2EleDuration),
			AffectedStat: attributes.PyroP,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})

		// Corresponding elemental DMG bonus
		if otherEle != attributes.Pyro {
			m2 := make([]float64, attributes.EndStatType)
			switch otherEle {
			case attributes.Hydro:
				m2[attributes.HydroP] = c2DmgBonus
			case attributes.Cryo:
				m2[attributes.CryoP] = c2DmgBonus
			case attributes.Electro:
				m2[attributes.ElectroP] = c2DmgBonus
			case attributes.Anemo:
				m2[attributes.AnemoP] = c2DmgBonus
			case attributes.Geo:
				m2[attributes.GeoP] = c2DmgBonus
			case attributes.Dendro:
				m2[attributes.DendroP] = c2DmgBonus
			}
			char.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag(c2EleBuffKey+"-"+otherEle.String(), c2EleDuration),
				AffectedStat: attributes.NoStat,
				Amount: func() ([]float64, bool) {
					return m2, true
				},
			})
		}
	}

	c.Core.Log.NewEvent("C2: Elemental DMG bonus applied to party", glog.LogCharacterEvent, c.Index).
		Write("pyro_bonus", c2DmgBonus).
		Write("other_element", otherEle.String())
}

// C4: Emanare's Source
// Durin's Elemental Burst DMG is increased by 40%
// Additionally, 30% chance to not consume Cycle of Enlightenment stacks (handled in c1OnEnemyHit)
func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}

	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = c4BurstDmgBonus

	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("durin-c4-burst-dmg", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagElementalBurst {
				return nil, false
			}
			return m, true
		},
	})
}

// C6: Dual Birth
// Principle of Purity: Burst DMG ignores 30% DEF, Dragon of White Flame reduces enemy DEF by 30% on hit (6 seconds)
// Principle of Darkness: Burst DMG ignores 70% DEF (30% base + 40% additional)
// Note: Purity uses "DEF ignore + DEF shred" combination, Darkness uses only "high-rate DEF ignore"
func (c *char) c6DragonWhiteFlameCB(a combat.AttackCB) {
	if c.Base.Cons < 6 {
		return
	}

	e, ok := a.Target.(*enemy.Enemy)
	if !ok {
		return
	}

	// Apply DEF shred
	e.AddDefMod(combat.DefMod{
		Base:  modifier.NewBaseWithHitlag(c6DefShredKey, c6DefShredDuration),
		Value: -c6DefShred,
	})

	c.Core.Log.NewEvent("C6: Dragon of White Flame DEF shred applied", glog.LogCharacterEvent, c.Index).
		Write("def_shred", c6DefShred)
}
