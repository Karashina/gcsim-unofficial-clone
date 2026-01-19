package columbina

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	// A0: Moonsign Benediction
	moonsignKey   = "moonsign"
	lbKeyStatus   = "LB-key"
	lcKeyStatus   = "LC-key"
	lcrsKeyStatus = "LCrs-key"

	// A1: Lunacy
	lunacyKey      = "lunacy"
	lunacyMaxStack = 3
	lunacyDur      = 10 * 60 // 10 seconds
	lunacyCRBonus  = 0.05    // 5% per stack

	// A4: Law of the New Moon
	a4Key = "law-of-new-moon"
)

// A0: Moonsign Benediction
// - Sets "moonsign" and "lcrs-key" status when Columbina is in party
// - Converts Electro-Charged → Lunar-Charged, Bloom → Lunar-Bloom, Hydro Crystallize → Lunar-Crystallize
// - Base DMG Bonus = 0.2% per 1000 Max HP, max 7%
func (c *char) a0Init() {
	for _, char := range c.Core.Player.Chars() {
		char.AddStatus(moonsignKey, -1, false)
		char.AddStatus(lbKeyStatus, -1, false)
		char.AddStatus(lcKeyStatus, -1, false)
		char.AddStatus(lcrsKeyStatus, -1, false)
	}

	maxval := 0.07
	val := min(maxval, c.MaxHP()/1000*0.002)

	for _, char := range c.Core.Player.Chars() {
		char.AddLCBaseReactBonusMod(character.LCBaseReactBonusMod{
			Base: modifier.NewBase("Moonlight, Lent Unto You (A0/LC)", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				return val, false
			},
		})
		char.AddLBBaseReactBonusMod(character.LBBaseReactBonusMod{
			Base: modifier.NewBase("Moonlight, Lent Unto You (A0/LB)", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				return val, false
			},
		})
		char.AddLCrsBaseReactBonusMod(character.LCrsBaseReactBonusMod{
			Base: modifier.NewBase("Moonlight, Lent Unto You (A0/LCrs)", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				return val, false
			},
		})
	}
}

// A1: Lunacy - CRIT Rate bonus based on stacks
func (c *char) a1Init() {
	if c.Base.Ascension < 1 {
		return
	}

	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase(lunacyKey+"-crit", -1),
		AffectedStat: attributes.CR,
		Amount: func() ([]float64, bool) {
			if c.lunacyStacks <= 0 {
				return nil, false
			}
			m := make([]float64, attributes.EndStatType)
			m[attributes.CR] = float64(c.lunacyStacks) * lunacyCRBonus
			return m, true
		},
	})
}

// a1OnGravityInterference adds Lunacy stack on Gravity Interference trigger
func (c *char) a1OnGravityInterference() {
	if c.Base.Ascension < 1 {
		return
	}

	// Add stack (max 3)
	c.lunacyStacks++
	if c.lunacyStacks > lunacyMaxStack {
		c.lunacyStacks = lunacyMaxStack
	}

	// Refresh duration
	c.AddStatus(lunacyKey, lunacyDur, true)

	c.Core.Log.NewEvent("Lunacy stack gained", glog.LogCharacterEvent, c.Index).
		Write("stacks", c.lunacyStacks).
		Write("crit_bonus", float64(c.lunacyStacks)*lunacyCRBonus)

	// Schedule stack decay
	c.lunacySrc = c.Core.F
	c.Core.Tasks.Add(c.lunacyDecay(c.Core.F), lunacyDur)
}

// lunacyDecay removes Lunacy stacks after duration
func (c *char) lunacyDecay(src int) func() {
	return func() {
		if c.lunacySrc != src {
			return
		}
		c.lunacyStacks = 0
		c.Core.Log.NewEvent("Lunacy expired", glog.LogCharacterEvent, c.Index)
	}
}
