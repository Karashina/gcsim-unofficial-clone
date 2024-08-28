package obsidiancodex

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	buffIcdKey = "obsidian-icd"
)

func init() {
	core.RegisterSetFunc(keys.ObsidianCodex, NewSet)
}

type Set struct {
	Index int
	Count int
}

func (s *Set) SetIndex(idx int) { s.Index = idx }
func (s *Set) GetCount() int    { return s.Count }
func (s *Set) Init() error      { return nil }

func NewSet(c *core.Core, char *character.CharWrapper, count int, param map[string]int) (info.Set, error) {
	s := Set{Count: count}

	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = 0.15

	if count >= 2 {
		c.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
			atk := args[1].(*combat.AttackEvent)
			if c.Player.Active() != char.Index {
				return false
			}
			if char.Index != atk.Info.ActorIndex {
				return false
			}

			char.AddAttackMod(character.AttackMod{
				Base: modifier.NewBaseWithHitlag("obsidiancodex-2pc", -1),
				Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
					switch char.OnNightsoul {
					case true:
						return m, true
					default:
						return nil, false
					}
				},
			})

			return false
		}, "obsidiancodex-2pc")
	}
	if count >= 4 {
		n := make([]float64, attributes.EndStatType)
		n[attributes.CR] = 0.4
		c.Events.Subscribe(event.OnNightsoulChange, func(args ...interface{}) bool {
			index := args[0].(int)
			amount := args[1].(float64)
			if char.StatusIsActive(buffIcdKey) {
				return false
			}
			if c.Player.Active() != char.Index {
				return false
			}
			if char.Index != index {
				return false
			}
			if amount < 1 {
				return false
			}

			char.AddAttackMod(character.AttackMod{
				Base: modifier.NewBaseWithHitlag("obsidiancodex-4pc", 6*60),
				Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
					return m, true
				},
			})
			char.AddStatus(buffIcdKey, 60, true)

			return false
		}, "obsidiancodex-4pc")
	}

	return &s, nil
}
