package gestofthemightywolf

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
	core.RegisterWeaponFunc(keys.GestOfTheMightyWolf, NewWeapon)
}

type Weapon struct {
	Index int

	stacks   int
	stackSrc int // source frame for stack expiry
	core     *core.Core
	char     *character.CharWrapper

	// Passive values
	dmgPerStack  float64
	cdmgPerStack float64
	hasHexBonus  bool
}

const (
	buffKey  = "gest-wolf-hymn"
	buffDur  = 4 * 60 // 4 seconds
	stackICD = 1      // 0.01s = ~1 frame ICD
)

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error {
	// Check Hexerei bonus at init time (after party is assembled)
	w.hasHexBonus = false
	hexereiCount := 0
	for _, char := range w.core.Player.Chars() {
		if result, err := char.Condition([]string{"hexerei"}); err == nil {
			if isHex, ok := result.(bool); ok && isHex {
				hexereiCount++
			}
		}
	}
	w.hasHexBonus = hexereiCount >= 2
	return nil
}

// Increase ATK SPD by 10%.
// Every time the equipping character's Normal Attack(s) hit opponent(s),
// when they cast their Elemental Skill, or when they begin their Charged Attack(s),
// gain 1/2/2 stacks of Four Winds' Hymn respectively:
// DMG dealt is increased by 7.5%/9.5%/11.5%/13.5%/15.5% for 4s. Max 4 stacks.
// This effect can be triggered once every 0.01s.
// Additionally, when the party has "Hexerei: Secret Rite", each stack of Four Winds' Hymn
// also increases the equipping character's CRIT DMG by 7.5%/9.5%/11.5%/13.5%/15.5%.
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		core: c,
		char: char,
	}
	r := p.Refine

	w.dmgPerStack = 0.055 + 0.02*float64(r) // R1=7.5%, R2=9.5%, R3=11.5%, R4=13.5%, R5=15.5%
	w.cdmgPerStack = w.dmgPerStack          // Same values for CRIT DMG

	// Permanent ATK SPD +10% (always active, does not scale with refine based on description)
	// Note: ATK SPD is typically a StatMod on AtkSpd
	atkSpd := make([]float64, attributes.EndStatType)
	atkSpd[attributes.AtkSpd] = 0.10
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("gest-wolf-atkspd", -1),
		AffectedStat: attributes.AtkSpd,
		Amount: func() ([]float64, bool) {
			return atkSpd, true
		},
	})

	// Subscribe to Normal Attack hits for 1 stack
	c.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagNormal {
			return false
		}
		w.addStacks(1)
		return false
	}, fmt.Sprintf("gest-wolf-na-%v", char.Base.Key.String()))

	// Subscribe to Skill cast for 2 stacks (on damage, not on cast)
	c.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagElementalArt && atk.Info.AttackTag != attacks.AttackTagElementalArtHold {
			return false
		}
		w.addStacks(2)
		return false
	}, fmt.Sprintf("gest-wolf-skill-%v", char.Base.Key.String()))

	// Subscribe to Charged Attack start for 2 stacks (on damage)
	c.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagExtra {
			return false
		}
		w.addStacks(2)
		return false
	}, fmt.Sprintf("gest-wolf-ca-%v", char.Base.Key.String()))

	// Apply DMG% buff via AttackMod (dynamic based on current stacks)
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("gest-wolf-dmg", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if !char.StatusIsActive(buffKey) {
				return nil, false
			}
			m := make([]float64, attributes.EndStatType)
			m[attributes.DmgP] = w.dmgPerStack * float64(w.stacks)
			if w.hasHexBonus {
				m[attributes.CD] = w.cdmgPerStack * float64(w.stacks)
			}
			return m, true
		},
	})

	return w, nil
}

// addStacks adds stacks to Four Winds' Hymn buff
func (w *Weapon) addStacks(count int) {
	// ICD check: 0.01s
	if w.char.StatusIsActive("gest-wolf-stack-icd") {
		return
	}
	w.char.AddStatus("gest-wolf-stack-icd", stackICD, true)

	// If buff expired, reset stacks
	if !w.char.StatusIsActive(buffKey) {
		w.stacks = 0
	}

	w.stacks += count
	if w.stacks > 4 {
		w.stacks = 4
	}

	// Refresh duration
	w.char.AddStatus(buffKey, buffDur, true)

	w.core.Log.NewEvent("gest-wolf stacks updated", glog.LogWeaponEvent, w.char.Index).
		Write("stacks", w.stacks).
		Write("count_added", count).
		Write("hex_bonus", w.hasHexBonus)
}
