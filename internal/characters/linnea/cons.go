package linnea

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// c1Init initializes Constellation 1:
// Gain 6 "Field Catalog" stacks (max 18) on Elemental Skill use and Lunar-Crystallize reactions.
// When a nearby party member deals LCrs DMG, consume 1 stack to add DEF×75% flat damage (LCrsFlatBonusMod).
// For Million Ton Crush, consume up to 5 stacks, adding DEF×150% flat damage per stack.
func (c *char) c1Init() {
	// Subscribe to base Moondrift Harmony: first moondrift projectile hit (AttackTagLCrsDamage + abil "lunar-crystallize")
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		ae, ok := args[1].(*combat.AttackEvent)
		if !ok {
			return false
		}
		if ae.Info.AttackTag != attacks.AttackTagLCrsDamage {
			return false
		}
		if ae.Info.Abil != "lunar-crystallize" {
			return false
		}
		c.onMoondriftHarmony()
		return false
	}, "linnea-moondrift-harmony")

	// Add LCrsFlatBonusMod to all party members.
	// When any party member (including Linnea herself) deals LCrs DMG,
	// consume 1 Field Catalog stack to add DEF×75% flat damage (C6: 2 stacks, DEF×112.5%).
	for _, char := range c.Core.Player.Chars() {
		char.AddLCrsFlatBonusMod(character.LCrsFlatBonusMod{
			Base: modifier.NewBase("linnea-c1-lcrs-flat", -1),
			Amount: func(atk combat.AttackInfo) (float64, bool) {
				return c.c1LCrsDamageBonus(), false
			},
		})
	}

	c.Core.Log.NewEvent("Linnea C1 active: Field Catalog system initialized", glog.LogCharacterEvent, c.Index)
}

// c1OnSkillUse handles C1 on Skill use (called from summonLumi)
func (c *char) c1OnSkillUse() {
	c.addFieldCatalogStacks(6)
}

// c1OnHarmony handles C1 on Moondrift Harmony
func (c *char) c1OnHarmony() {
	c.addFieldCatalogStacks(6)
}

// addFieldCatalogStacks adds Field Catalog stacks
func (c *char) addFieldCatalogStacks(n int) {
	maxStacks := maxFieldCatalog
	if c.Base.Cons >= 6 {
		// C6: instantly fill to max stacks on trigger
		c.fieldCatalogStacks = maxStacks
	} else {
		c.fieldCatalogStacks = min(maxStacks, c.fieldCatalogStacks+n)
	}
	c.fieldCatalogSrc = c.Core.F
	c.AddStatus(fieldCatalogKey, fieldCatalogDuration, true)

	c.Core.Log.NewEvent("Field Catalog stacks updated", glog.LogCharacterEvent, c.Index).
		Write("stacks", c.fieldCatalogStacks)
}

// c1LCrsDamageBonus returns the Field Catalog consumption bonus on LCrs DMG.
// consume 1 stack to add DEF×75% flat damage
func (c *char) c1LCrsDamageBonus() float64 {
	if c.fieldCatalogStacks <= 0 || !c.StatusIsActive(fieldCatalogKey) {
		return 0
	}

	consume := 1
	bonusMult := 0.75
	if c.Base.Cons >= 6 {
		// C6: consume twice (2 stacks), DMG boosted to 150% of original (DEF*75%*1.5 = DEF*112.5%)
		consume = 2
		bonusMult = 1.125
	}
	if c.fieldCatalogStacks < consume {
		consume = c.fieldCatalogStacks
		if c.Base.Cons >= 6 && consume < 2 {
			bonusMult = 0.75 // not enough stacks for full C6: fall back to base rate
		}
	}

	c.fieldCatalogStacks -= consume
	bonus := c.TotalDef(false) * bonusMult

	c.Core.Log.NewEvent("Field Catalog consumed for LCrs damage", glog.LogCharacterEvent, c.Index).
		Write("consumed", consume).
		Write("bonus", bonus).
		Write("remaining", c.fieldCatalogStacks)

	return bonus
}

// c1MillionTonCrushBonus returns the Field Catalog consumption bonus for Million Ton Crush.
// consume up to 5 stacks, adding DEF×150% flat damage per stack
func (c *char) c1MillionTonCrushBonus() float64 {
	if c.fieldCatalogStacks <= 0 || !c.StatusIsActive(fieldCatalogKey) {
		return 0
	}

	maxConsume := 5
	bonusPerStack := 1.50
	if c.Base.Cons >= 6 {
		// C6: consume twice (up to 10 stacks), DMG boosted to 150% of original (DEF*150%*1.5 = DEF*225% per pair)
		maxConsume = 10
		bonusPerStack = 2.25
	}

	consume := min(maxConsume, c.fieldCatalogStacks)
	effectiveStacks := consume
	if c.Base.Cons >= 6 {
		effectiveStacks = consume / 2
		if effectiveStacks == 0 {
			effectiveStacks = 1
		}
	}

	c.fieldCatalogStacks -= consume
	bonus := c.TotalDef(false) * bonusPerStack * float64(effectiveStacks)

	c.Core.Log.NewEvent("Field Catalog consumed for Million Ton Crush", glog.LogCharacterEvent, c.Index).
		Write("consumed", consume).
		Write("effectiveStacks", effectiveStacks).
		Write("bonus", bonus).
		Write("remaining", c.fieldCatalogStacks)

	return bonus
}

// c2Init initializes Constellation 2:
// After Moondrift Harmony, Hydro/Geo party members' CRIT DMG increases by 40% for 8 seconds.
// Million Ton Crush CRIT DMG increases by 150%.
// When Moonsign: Ascendant, HOH and MTC also trigger Moondrift Harmony.
func (c *char) c2Init() {
	// C2 Ascendant Gleam: HOH/MTC trigger Moondrift Harmony (C1/C2/C4 effects)
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		if !c.MoonsignAscendant {
			return false
		}
		ae, ok := args[1].(*combat.AttackEvent)
		if !ok {
			return false
		}
		if ae.Info.ActorIndex != c.Index {
			return false
		}
		abil := ae.Info.Abil
		if abil != "Lumi Heavy Overdrive Hammer (Lunar-Crystallize)" &&
			abil != "Lumi Million Ton Crush (Lunar-Crystallize)" {
			return false
		}
		c.onMoondriftHarmony()
		return false
	}, "linnea-ascendant-harmony")

	c.Core.Log.NewEvent("Linnea C2 active: Moondrift CRIT DMG bonus initialized", glog.LogCharacterEvent, c.Index)
}

// c2OnHarmony handles C2 on Moondrift Harmony:
// grants Hydro/Geo party members CRIT DMG +40% for 8 seconds.
func (c *char) c2OnHarmony() {
	const c2Duration = 8 * 60

	for _, char := range c.Core.Player.Chars() {
		ele := char.Base.Element
		if ele != attributes.Hydro && ele != attributes.Geo {
			continue
		}

		idx := char.Index
		cdMod := make([]float64, attributes.EndStatType)
		cdMod[attributes.CD] = 0.40

		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(c2CritDmgKey, c2Duration),
			AffectedStat: attributes.CD,
			Amount: func() ([]float64, bool) {
				_ = idx // capture for closure
				return cdMod, true
			},
		})
	}

	c.Core.Log.NewEvent("Linnea C2: Hydro/Geo party CRIT DMG +40%", glog.LogCharacterEvent, c.Index).
		Write("duration", c2Duration)
}

// c4Init initializes Constellation 4:
// After Moondrift Harmony, Linnea's and the active character's DEF increases by 25% for 5 seconds.
func (c *char) c4Init() {
	c.Core.Log.NewEvent("Linnea C4 active: DEF buff on Moondrift Harmony", glog.LogCharacterEvent, c.Index)
}

// c4OnHarmony handles C4 on Moondrift Harmony
func (c *char) c4OnHarmony() {
	const c4Duration = 5 * 60

	defMod := make([]float64, attributes.EndStatType)
	defMod[attributes.DEFP] = 0.25

	// DEF +25% for Linnea herself
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag(c4DefKey, c4Duration),
		AffectedStat: attributes.DEFP,
		Amount: func() ([]float64, bool) {
			return defMod, true
		},
	})

	// DEF +25% for active character (stacks if Linnea is active)
	activeIdx := c.Core.Player.Active()
	activeChar := c.Core.Player.ByIndex(activeIdx)
	activeDefMod := make([]float64, attributes.EndStatType)
	activeDefMod[attributes.DEFP] = 0.25
	activeChar.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag(c4DefActiveKey, c4Duration),
		AffectedStat: attributes.DEFP,
		Amount: func() ([]float64, bool) {
			return activeDefMod, true
		},
	})

	c.Core.Log.NewEvent("Linnea C4: DEF +25% applied", glog.LogCharacterEvent, c.Index).
		Write("duration", c4Duration)
}

// c6Init initializes Constellation 6:
// Field Catalog stack gain is enhanced; stacks are instantly filled to max on trigger.
// Stack consumption doubles and the bonus increases to 150%.
// When Moonsign: Ascendant, LCrs DMG Elevation increases by 25%.
func (c *char) c6Init() {
	// Elevation +25% for LCrs DMG when Moonsign: Ascendant
	c.AddElevationMod(character.ElevationMod{
		Base: modifier.NewBase("linnea-c6-elevation", -1),
		Amount: func(ai combat.AttackInfo) (float64, bool) {
			if !c.MoonsignAscendant {
				return 0, false
			}
			if ai.AttackTag != attacks.AttackTagLCrsDamage {
				return 0, false
			}
			// C6: applies to all nearby party members' LCrs DMG
			return 0.25, false
		},
	})

	c.Core.Log.NewEvent("Linnea C6 active: Enhanced Field Catalog and Elevation bonus", glog.LogCharacterEvent, c.Index)
}
