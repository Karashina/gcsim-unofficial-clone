package zibai

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/reactions"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// a0Init initializes the Moonsign Benediction passive
// When a party member triggers a Hydro Crystallize reaction, it will be converted
// into the Lunar-Crystallize reaction, with every 100 DEF that Zibai has increasing
// Lunar-Crystallize's Base DMG by 0.7%, up to a maximum of 14%.
// Additionally, when Zibai is in the party, the party's Moonsign will increase by 1 level.
func (c *char) a0Init() {
	// Grant LCrs-Key to all party members (enables Lunar-Crystallize)
	for _, char := range c.Core.Player.Chars() {
		char.AddStatus("LCrs-key", -1, true)
	}
	// Grant moonsignKey to Zibai (increases party Moonsign level by 1)
	// This is counted during party initialization to determine Moonsign state
	c.AddStatus("moonsignKey", -1, true)
	// Add Lunar-Crystallize Base DMG bonus based on DEF
	// STUB: This should modify the Lunar-Crystallize reaction damage calculation
	// Every 100 DEF = 0.7% bonus, max 14% (2000 DEF)
	c.AddLCrsBaseReactBonusMod(character.LCrsBaseReactBonusMod{
		Base: modifier.NewBase("the-coursing-sun-and-moon-a0", -1),
		Amount: func(ai combat.AttackInfo) (float64, bool) {
			maxval := 0.14
			val := min(maxval, c.TotalDef(false)/100*0.007)
			return val, false
		},
	})
}

// a1Init initializes The Selenic Adeptus Descends passive
// When casting the Elemental Skill, or when nearby party members trigger Lunar-Crystallize reaction dmg,
// Zibai gains the Selenic Descent effect for 4s: The DMG dealt by the 2nd hit of Spirit Steed's Stride
// is increased by 60% of Zibai's DEF.
func (c *char) a1Init() {
	if c.Base.Ascension < 1 {
		return
	}
	const selenicDescentDuration = 4 * 60 // 4 seconds

	// Subscribe to Skill cast
	c.Core.Events.Subscribe(event.OnSkill, func(args ...interface{}) bool {
		if c.Core.Player.Active() != c.Index {
			return false
		}
		c.applySelenicDescent(selenicDescentDuration)
		return false
	}, "zibai-a1-skill")

	// Subscribe to Lunar-Crystallize reaction damage from party members
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		ae, ok := args[1].(*combat.AttackEvent)
		if !ok {
			return false
		}
		// Only trigger on Lunar-Crystallize reaction damage
		if ae.Info.Abil != string(reactions.LunarCrystallize) {
			return false
		}
		c.applySelenicDescent(selenicDescentDuration)
		return false
	}, "zibai-a1-lcrs")
}

// applySelenicDescent applies the Selenic Descent buff
func (c *char) applySelenicDescent(duration int) {
	c.AddStatus(selenicDescentKey, duration, true)

	c.Core.Log.NewEvent("Zibai gains Selenic Descent", glog.LogCharacterEvent, c.Index).
		Write("duration", duration)
}

// a4Init initializes Layered Peaks Pierce the Clouds passive
// Other Geo party members increase Zibai's DEF by 15% each.
// Hydro party members increase her Elemental Mastery by 60 each.
func (c *char) a4Init() {
	if c.Base.Ascension < 4 {
		return
	}

	geoCount := 0
	hydroCount := 0

	for _, char := range c.Core.Player.Chars() {
		if char.Index == c.Index {
			continue // Skip self
		}
		switch char.Base.Element {
		case attributes.Geo:
			geoCount++
		case attributes.Hydro:
			hydroCount++
		}
	}

	defBonus := 0.15 * float64(geoCount)
	emBonus := 60.0 * float64(hydroCount)

	if defBonus > 0 {
		defMod := make([]float64, attributes.EndStatType)
		defMod[attributes.DEFP] = defBonus
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("zibai-a4-defp", -1),
			AffectedStat: attributes.DEFP,
			Amount: func() ([]float64, bool) {
				return defMod, true
			},
		})
	}

	if emBonus > 0 {
		emMod := make([]float64, attributes.EndStatType)
		emMod[attributes.EM] = emBonus
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("zibai-a4-em", -1),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				return emMod, true
			},
		})
	}

	if defBonus > 0 || emBonus > 0 {
		c.Core.Log.NewEvent("Zibai A4 stat bonus applied", glog.LogCharacterEvent, c.Index).
			Write("geo_count", geoCount).
			Write("hydro_count", hydroCount).
			Write("def_bonus", defBonus).
			Write("em_bonus", emBonus)
	}
}
