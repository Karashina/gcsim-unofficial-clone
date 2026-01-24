package nocturnescurtaincall

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.NocturnesCurtainCall, NewWeapon)
}

type Weapon struct {
	Index int
	core  *core.Core
	char  *character.CharWrapper
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// Nocturne's Curtain Call
// Max HP increases by 10/12/14/16/18%.
// When triggering Lunar reactions or inflicting Lunar Reaction DMG on opponents,
// the equipping character will recover 14/15/16/17/18 Energy, and receive the
// Bountiful Sea's Sacred Wine effect for 12s:
// Max HP increases by an additional 14/16/18/20/22%,
// CRIT DMG from Lunar Reaction DMG increases by 60/80/100/120/140%.
// The Energy recovery effect can be triggered at most once every 18s,
// and can be triggered even when the equipping character is off-field.

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		core: c,
		char: char,
	}
	r := p.Refine

	// Base HP increase (10/12/14/16/18%)
	baseHPBonus := 0.08 + 0.02*float64(r)
	val := make([]float64, attributes.EndStatType)
	val[attributes.HPP] = baseHPBonus

	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("nocturnes-curtain-call-base", -1),
		AffectedStat: attributes.HPP,
		Amount: func() ([]float64, bool) {
			return val, true
		},
	})

	// Energy recovery and buff on Lunar reaction
	energyAmt := []float64{14, 15, 16, 17, 18}
	buffHPBonus := []float64{0.14, 0.16, 0.18, 0.20, 0.22}
	buffCDBonus := []float64{0.60, 0.80, 1.00, 1.20, 1.40}

	const buffKey = "nocturnes-curtain-call-buff"
	const icdKey = "nocturnes-curtain-call-icd"
	buffDuration := 720 // 12s
	icdDuration := 1080 // 18s

	// Subscribe to Lunar reaction events
	lunarReactionTrigger := func(args ...interface{}) bool {
		// args[0] should be the reactable target
		// Check if the attack actor is this character
		atk, ok := args[1].(*combat.AttackEvent)
		if !ok {
			return false
		}
		if atk.Info.ActorIndex != char.Index {
			return false
		}

		// Check ICD
		if !char.StatusIsActive(icdKey) {
			// Recover energy
			char.AddEnergy("nocturnes-curtain-call", energyAmt[r-1])
			c.Log.NewEvent("nocturnes curtain call energy recovery", glog.LogWeaponEvent, char.Index).
				Write("energy", energyAmt[r-1])
				// Add ICD
			char.AddStatus(icdKey, icdDuration, true)
		}

		// Apply buff (HP + CRIT DMG for Lunar reactions)
		char.AddStatus(buffKey, buffDuration, true)

		return false
	}

	c.Events.Subscribe(event.OnLunarCharged, lunarReactionTrigger, fmt.Sprintf("nocturnes-curtain-call-lc-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnLunarBloom, lunarReactionTrigger, fmt.Sprintf("nocturnes-curtain-call-lb-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnLunarCrystallize, lunarReactionTrigger, fmt.Sprintf("nocturnes-curtain-call-lcrs-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk, ok := args[1].(*combat.AttackEvent)
		if !ok {
			return false
		}
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagLCDamage &&
			atk.Info.AttackTag != attacks.AttackTagLBDamage &&
			atk.Info.AttackTag != attacks.AttackTagLCrsDamage {
			return false
		}

		// Check ICD
		if !char.StatusIsActive(icdKey) {
			// Recover energy
			char.AddEnergy("nocturnes-curtain-call", energyAmt[r-1])
			c.Log.NewEvent("nocturnes curtain call energy recovery", glog.LogWeaponEvent, char.Index).
				Write("energy", energyAmt[r-1])
				// Add ICD
			char.AddStatus(icdKey, icdDuration, true)
		}

		// Apply buff (HP + CRIT DMG for Lunar reactions)
		char.AddStatus(buffKey, buffDuration, true)

		return false
	}, "nocturnes-curtain-call-enemy-dmg-"+char.Base.Key.String())

	// HP bonus from buff
	buffVal := make([]float64, attributes.EndStatType)
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase(fmt.Sprintf("%s-hp", buffKey), -1),
		AffectedStat: attributes.HPP,
		Amount: func() ([]float64, bool) {
			if !char.StatusIsActive(buffKey) {
				return nil, false
			}
			buffVal[attributes.HPP] = buffHPBonus[r-1]
			return buffVal, true
		},
	})

	// CRIT DMG bonus for Lunar Reaction DMG
	cdVal := make([]float64, attributes.EndStatType)
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase(fmt.Sprintf("%s-cd", buffKey), -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if !char.StatusIsActive(buffKey) {
				return nil, false
			}
			// Check if this is Lunar Reaction DMG
			if atk.Info.AttackTag != attacks.AttackTagLCDamage &&
				atk.Info.AttackTag != attacks.AttackTagLBDamage &&
				atk.Info.AttackTag != attacks.AttackTagLCrsDamage {
				return nil, false
			}
			cdVal[attributes.CD] = buffCDBonus[r-1]
			return cdVal, true
		},
	})

	return w, nil
}
