package adaycarvedfromrisingwinds

import (
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
	core.RegisterSetFunc(keys.ADayCarvedFromRisingWinds, NewSet)
}

type Set struct {
	Index int
	Count int
}

func (s *Set) SetIndex(idx int) { s.Index = idx }
func (s *Set) GetCount() int    { return s.Count }
func (s *Set) Init() error      { return nil }

// A Day Carved From Rising Winds
// 2-set: ATK +18%.
// 4-set: After a Normal Attack, Charged Attack, Elemental Skill or Elemental Burst
// hits an opponent, gain the Blessing of Pastoral Winds effect for 6s: ATK is increased by 25%.
// If the equipping character has completed Witch's Homework, Blessing of Pastoral Winds
// will be upgraded to Resolve of Pastoral Winds, which also increases the CRIT Rate
// of the equipping character by an additional 20%.
// This effect can be triggered even when the character is off-field.

func NewSet(c *core.Core, char *character.CharWrapper, count int, param map[string]int) (info.Set, error) {
	s := Set{Count: count}

	if count >= 2 {
		// 2pc: ATK +18%
		m := make([]float64, attributes.EndStatType)
		m[attributes.ATKP] = 0.18
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("adaycarved-2pc", -1),
			AffectedStat: attributes.ATKP,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}

	if count >= 4 {
		const buffKey = "adaycarved-4pc-buff"
		buffDuration := 360 // 6s

		// Subscribe to attack hit events
		c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
			atk := args[1].(*combat.AttackEvent)
			if atk.Info.ActorIndex != char.Index {
				return false
			}

			// Check if it's Normal, Charged, Skill, or Burst attack
			if atk.Info.AttackTag != attacks.AttackTagNormal &&
				atk.Info.AttackTag != attacks.AttackTagExtra &&
				atk.Info.AttackTag != attacks.AttackTagElementalArt &&
				atk.Info.AttackTag != attacks.AttackTagElementalBurst {
				return false
			}

			// Activate buff
			char.AddStatus(buffKey, buffDuration, true)

			c.Log.NewEvent("a day carved from rising winds 4pc triggered", glog.LogArtifactEvent, char.Index)

			return false
		}, "adaycarved-4pc")

		// ATK bonus from buff
		atkVal := make([]float64, attributes.EndStatType)
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("adaycarved-4pc-atk", -1),
			AffectedStat: attributes.ATKP,
			Amount: func() ([]float64, bool) {
				if !char.StatusIsActive(buffKey) {
					return nil, false
				}
				atkVal[attributes.ATKP] = 0.25
				return atkVal, true
			},
		})

		// CRIT Rate bonus if Witch's Homework is completed
		// Witch's Homework is a status that should be set by the simulation or character logic
		crVal := make([]float64, attributes.EndStatType)
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("adaycarved-4pc-cr", -1),
			AffectedStat: attributes.CR,
			Amount: func() ([]float64, bool) {
				if !char.StatusIsActive(buffKey) {
					return nil, false
				}
				// Check if Witch's Homework is completed
				if !char.StatusIsActive("witchs-homework") {
					return nil, false
				}
				crVal[attributes.CR] = 0.20
				return crVal, true
			},
		})
	}

	return &s, nil
}
