package linnea

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// a0Init initializes the moonsign passive.
// When a party member triggers a Hydro Crystallize reaction, it is converted to a Lunar-Crystallize reaction.
// For every 100 DEF Linnea has, Lunar-Crystallize base DMG increases by 0.7% (max 14%).
// The party's moonsign level increases by 1.
func (c *char) a0Init() {
	// apply LCrs key to all party members (enables Lunar-Crystallize)
	for _, char := range c.Core.Player.Chars() {
		char.AddStatus("LCrs-key", -1, true)
	}
	// apply moonsignKey (moonsign level +1)
	c.AddStatus("moonsignKey", -1, true)

	// add DEF-based Lunar-Crystallize base damage bonus
	// 0.7% per 100 DEF, max 14%
	c.AddLCrsBaseReactBonusMod(character.LCrsBaseReactBonusMod{
		Base: modifier.NewBase("linnea-a0-lcrs-base", -1),
		Amount: func(ai combat.AttackInfo) (float64, bool) {
			maxVal := 0.14
			val := min(maxVal, c.TotalDef(false)/100*0.007)
			return val, false
		},
	})
}

// a1Init initializes Passive Talent 1 "Field Observation Notes".
// While Lumi is on the field, all enemies' Geo RES is reduced by 15%.
// Moonsign: Ascendant Gleam: reduced by an additional 15% (total 30%).
func (c *char) a1Init() {
	if c.Base.Ascension < 1 {
		return
	}
	c.a1Tick()
}

// a1Tick applies Geo RES reduction to all enemies every 60f (hitlag-independent)
func (c *char) a1Tick() {
	if c.lumiActive {
		for _, t := range c.Core.Combat.Enemies() {
			e, ok := t.(*enemy.Enemy)
			if !ok {
				continue
			}
			// base Geo RES -15% (expires at 90f, maintained by 60f refresh cycle)
			e.AddResistMod(combat.ResistMod{
				Base:  modifier.NewBase(a1GeoResKey, 90),
				Ele:   attributes.Geo,
				Value: -0.15,
			})
			// Ascendant Gleam: additional -15%
			if c.MoonsignAscendant {
				e.AddResistMod(combat.ResistMod{
					Base:  modifier.NewBase(a1GeoResAscendKey, 90),
					Ele:   attributes.Geo,
					Value: -0.15,
				})
			}
		}
	}
	c.QueueCharTask(c.a1Tick, 60)
}

// a4Init initializes Passive Talent 4 "Encyclopedia of All Things".
// Grants EM buff based on the active character.
// Increases EM by 5% of Linnea's DEF.
// Moonsign character -> that character's EM increases.
// Non-Moonsign character -> Linnea's own EM increases.
func (c *char) a4Init() {
	if c.Base.Ascension < 4 {
		return
	}

	// add StatMod to each party member; active character is checked dynamically inside Amount
	for _, char := range c.Core.Player.Chars() {
		idx := char.Index
		emMod := make([]float64, attributes.EndStatType)
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("linnea-a4-em", -1),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				// do not apply if not the active character
				if c.Core.Player.Active() != idx {
					return nil, false
				}
				// dynamically check if active character has moonsign
				activeChar := c.Core.Player.ByIndex(idx)
				if !activeChar.StatusIsActive("moonsignKey") {
					return nil, false
				}
				emMod[attributes.EM] = c.TotalDef(false) * 0.05
				return emMod, true
			},
		})
	}

	// Linnea herself: EM increase when a non-Moonsign character is active
	emModSelf := make([]float64, attributes.EndStatType)
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("linnea-a4-em-self", -1),
		AffectedStat: attributes.EM,
		Amount: func() ([]float64, bool) {
			activeChar := c.Core.Player.ByIndex(c.Core.Player.Active())
			if activeChar.StatusIsActive("moonsignKey") {
				return nil, false
			}
			emModSelf[attributes.EM] = c.TotalDef(false) * 0.05
			return emModSelf, true
		},
	})

	c.Core.Log.NewEvent("Linnea A4 EM sharing initialized", glog.LogCharacterEvent, c.Index)
}
