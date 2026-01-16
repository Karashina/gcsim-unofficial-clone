package aubadeofmorningstarandmoon

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterSetFunc(keys.AubadeOfMorningstarAndMoon, NewSet)
}

type Set struct {
	Index int
	Count int
}

func (s *Set) SetIndex(idx int) { s.Index = idx }
func (s *Set) GetCount() int    { return s.Count }
func (s *Set) Init() error      { return nil }

// Aubade of Morningstar and Moon
// 2-set: Increases Elemental Mastery by 80.
// 4-set: When the equipping character is off-field, Lunar-Charged, Lunar-Bloom,
// and Lunar-Crystallize Bonus is increased by 20%.
// When the party's Moonsign Level is at least Ascendant Gleam, Lunar Reaction DMG
// will be further increased by 40%.
// This effect will disappear after the equipping character is active for 3s.

func NewSet(c *core.Core, char *character.CharWrapper, count int, param map[string]int) (info.Set, error) {
	s := Set{Count: count}

	if count >= 2 {
		// 2pc: EM +80
		m := make([]float64, attributes.EndStatType)
		m[attributes.EM] = 80
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("aubade-2pc", -1),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}

	if count >= 4 {
		const activeKey = "aubade-4pc-active-timer"
		activeTimeout := 180 // 3s

		// Track when character becomes active
		c.Events.Subscribe(event.OnCharacterSwap, func(args ...interface{}) bool {
			if c.Player.Active() != char.Index {
				return false
			}
			// Start timer when character becomes active
			char.AddStatus(activeKey, activeTimeout, true)
			return false
		}, "aubade-4pc-swap")

		// Add attack mod for Lunar Reaction DMG bonus
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("aubade-4pc", -1),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				// Check if this is Lunar Reaction DMG
				if atk.Info.AttackTag != attacks.AttackTagLCDamage &&
					atk.Info.AttackTag != attacks.AttackTagLBDamage &&
					atk.Info.AttackTag != attacks.AttackTagLCrsDamage {
					return nil, false
				}

				// Check if character is off-field (active timer not active)
				if char.StatusIsActive(activeKey) {
					return nil, false
				}

				val := make([]float64, attributes.EndStatType)

				// Base bonus: 20% when off-field
				bonus := 0.20

				// Additional bonus: 40% if Moonsign Level is at least Ascendant Gleam
				// Check if character has Moonsign Ascendant status
				if char.MoonsignAscendant {
					bonus += 0.40
				}

				val[attributes.DmgP] = bonus
				return val, true
			},
		})
	}

	return &s, nil
}
