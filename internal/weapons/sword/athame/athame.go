package athame

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
	core.RegisterWeaponFunc(keys.AthameArtis, NewWeapon)
}

type Weapon struct {
	Index int
	c     *core.Core
	char  *character.CharWrapper
}

const (
	bladeKey      = "athame-blade"
	bladeDuration = 3 * 60 // 3s
)

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		c:    c,
		char: char,
	}
	r := p.Refine

	// CRIT DMG bonus for Elemental Burst
	burstCritDmg := 0.14 + float64(r)*0.04 // 16/20/24/28/32%
	m := make([]float64, attributes.EndStatType)
	m[attributes.CD] = burstCritDmg
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("athame-burst-cd", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagElementalBurst {
				return nil, false
			}
			return m, true
		},
	})

	// Blade of the Daylight Hours effect on Burst hit
	atkBonus := 0.18 + float64(r)*0.05      // 20/25/30/35/40%
	partyAtkBonus := 0.14 + float64(r)*0.04 // 16/20/24/28/32%

	c.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagElementalBurst {
			return false
		}

		// Check for Hexerei: Secret Rite (2+ Hexerei characters)
		hasHexerei := w.countHexereiCharacters() >= 2
		multiplier := 1.0
		if hasHexerei {
			multiplier = 1.75 // 75% increase
		}

		// Apply ATK bonus to wielder
		wielderBonus := atkBonus * multiplier
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(bladeKey, bladeDuration),
			AffectedStat: attributes.ATKP,
			Amount: func() ([]float64, bool) {
				return w.makeStatArray(attributes.ATKP, wielderBonus), true
			},
		})

		// Apply ATK bonus to nearby active party members
		partyBonus := partyAtkBonus * multiplier
		for _, other := range c.Player.Chars() {
			if other.Index == char.Index {
				continue
			}
			other.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag(bladeKey+fmt.Sprintf("-%v", other.Index), bladeDuration),
				AffectedStat: attributes.ATKP,
				Amount: func() ([]float64, bool) {
					return w.makeStatArray(attributes.ATKP, partyBonus), true
				},
			})
		}

		c.Log.NewEvent("Athame Artis: Blade of the Daylight Hours triggered", glog.LogWeaponEvent, char.Index).
			Write("wielder_atk_bonus", wielderBonus).
			Write("party_atk_bonus", partyBonus).
			Write("has_hexerei", hasHexerei)

		return false
	}, fmt.Sprintf("athame-%v", char.Base.Key.String()))

	return w, nil
}

func (w *Weapon) countHexereiCharacters() int {
	count := 0
	for _, char := range w.c.Player.Chars() {
		// Check if character has Hexerei trait
		// This assumes characters have a method or field to identify Hexerei status
		// For now, we'll check if they have any "hexerei" status
		if char.StatusIsActive("hexerei-character") {
			count++
		}
	}
	return count
}

func (w *Weapon) makeStatArray(stat attributes.Stat, value float64) []float64 {
	stats := make([]float64, attributes.EndStatType)
	stats[stat] = value
	return stats
}
