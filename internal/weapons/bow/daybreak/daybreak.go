package daybreak

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
	core.RegisterWeaponFunc(keys.TheDaybreakChronicles, NewWeapon)
}

type Weapon struct {
	Index int
	c     *core.Core
	char  *character.CharWrapper

	// Stirring Dawn Breeze tracking
	naDmgBonus    float64
	skillDmgBonus float64
	burstDmgBonus float64
	lastHitTime   int
}

const (
	hitICDFrames = 6 // 0.1s at 60fps
)

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		c:    c,
		char: char,
	}
	r := p.Refine

	maxBonusScaled := (0.45 + float64(r)*0.15)         // 60/75/90/105/120%
	decayRateScaled := (0.075 + float64(r)*0.025)      // 10/12.5/15/17.5/20%
	increaseAmountScaled := (0.075 + float64(r)*0.025) // 10/12.5/15/17.5/20%

	// Check for Hexerei: Secret Rite
	hasHexerei := w.countHexereiCharacters() >= 2
	hexereiIncrease := 0.0
	if hasHexerei {
		hexereiIncrease = 0.15 + float64(r)*0.05 // 20/25/30/35/40%
	}

	// Continuous decay (assume always in-combat)
	c.Tasks.Add(func() {
		// Decay all bonuses
		if w.naDmgBonus > 0 {
			w.naDmgBonus -= decayRateScaled / 60.0 // Per frame decay
			if w.naDmgBonus < 0 {
				w.naDmgBonus = 0
			}
		}
		if w.skillDmgBonus > 0 {
			w.skillDmgBonus -= decayRateScaled / 60.0
			if w.skillDmgBonus < 0 {
				w.skillDmgBonus = 0
			}
		}
		if w.burstDmgBonus > 0 {
			w.burstDmgBonus -= decayRateScaled / 60.0
			if w.burstDmgBonus < 0 {
				w.burstDmgBonus = 0
			}
		}
	}, 1)

	// On hit: increase corresponding DMG bonus
	c.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}

		// ICD check
		if c.F < w.lastHitTime+hitICDFrames {
			return false
		}

		attackTag := atk.Info.AttackTag
		var bonusPtr *float64
		var bonusName string

		switch {
		case attackTag == attacks.AttackTagNormal || attackTag == attacks.AttackTagExtra || attackTag == attacks.AttackTagPlunge:
			bonusPtr = &w.naDmgBonus
			bonusName = "NA"
		case attackTag == attacks.AttackTagElementalArt || attackTag == attacks.AttackTagElementalArtHold:
			bonusPtr = &w.skillDmgBonus
			bonusName = "Skill"
		case attackTag == attacks.AttackTagElementalBurst:
			bonusPtr = &w.burstDmgBonus
			bonusName = "Burst"
		default:
			return false
		}

		w.lastHitTime = c.F

		if hasHexerei {
			// Hexerei mode: increase all bonuses
			w.naDmgBonus += hexereiIncrease
			w.skillDmgBonus += hexereiIncrease
			w.burstDmgBonus += hexereiIncrease

			if w.naDmgBonus > maxBonusScaled {
				w.naDmgBonus = maxBonusScaled
			}
			if w.skillDmgBonus > maxBonusScaled {
				w.skillDmgBonus = maxBonusScaled
			}
			if w.burstDmgBonus > maxBonusScaled {
				w.burstDmgBonus = maxBonusScaled
			}

			c.Log.NewEvent("Daybreak Chronicles: Hexerei bonus applied", glog.LogWeaponEvent, char.Index).
				Write("na_bonus", w.naDmgBonus).
				Write("skill_bonus", w.skillDmgBonus).
				Write("burst_bonus", w.burstDmgBonus)
		} else {
			// Normal mode: increase only corresponding bonus
			*bonusPtr += increaseAmountScaled
			if *bonusPtr > maxBonusScaled {
				*bonusPtr = maxBonusScaled
			}

			c.Log.NewEvent("Daybreak Chronicles: DMG bonus increased", glog.LogWeaponEvent, char.Index).
				Write("attack_type", bonusName).
				Write("bonus", *bonusPtr)
		}

		return false
	}, fmt.Sprintf("daybreak-hit-%v", char.Base.Key.String()))

	// Apply DMG bonuses
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("daybreak-dmg", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			m := make([]float64, attributes.EndStatType)
			applied := false

			switch atk.Info.AttackTag {
			case attacks.AttackTagNormal, attacks.AttackTagExtra, attacks.AttackTagPlunge:
				if w.naDmgBonus > 0 {
					m[attributes.DmgP] = w.naDmgBonus
					applied = true
				}
			case attacks.AttackTagElementalArt, attacks.AttackTagElementalArtHold:
				if w.skillDmgBonus > 0 {
					m[attributes.DmgP] = w.skillDmgBonus
					applied = true
				}
			case attacks.AttackTagElementalBurst:
				if w.burstDmgBonus > 0 {
					m[attributes.DmgP] = w.burstDmgBonus
					applied = true
				}
			}

			return m, applied
		},
	})

	return w, nil
}

func (w *Weapon) countHexereiCharacters() int {
	count := 0
	for _, char := range w.c.Player.Chars() {
		if char.StatusIsActive("hexerei-character") {
			count++
		}
	}
	return count
}
