package finaleofthedeepgalleries

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterSetFunc(keys.FinaleOfTheDeepGalleries, NewSet)
}

type Set struct {
	Index int
	Count int
	c     *core.Core
	char  *character.CharWrapper
}

const (
	buffstopkeyna    = "fodg-4pc-stop-na"
	buffstopkeyburst = "fodg-4pc-stop-burst"
)

func (s *Set) SetIndex(idx int) { s.Index = idx }
func (s *Set) GetCount() int    { return s.Count }
func (s *Set) Init() error      { return nil }

func NewSet(c *core.Core, char *character.CharWrapper, count int, param map[string]int) (info.Set, error) {
	s := Set{
		Count: count,
		c:     c,
		char:  char,
	}
	if count >= 2 {
		m := make([]float64, attributes.EndStatType)
		m[attributes.CryoP] = 0.15
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("fodg-2pc", -1),
			AffectedStat: attributes.CryoP,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}
	if count >= 4 {
		nabuff := make([]float64, attributes.EndStatType)
		nabuff[attributes.DmgP] = 0.6
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("fodg-4pc-na", -1),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				if char.Energy > 0 {
					return nil, false
				}
				if char.StatusIsActive(buffstopkeyna) {
					return nil, false
				}
				if atk.Info.AttackTag == attacks.AttackTagNormal {
					return nabuff, true
				}
				return nil, false
			},
		})
		burstbuff := make([]float64, attributes.EndStatType)
		burstbuff[attributes.DmgP] = 0.6
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("fodg-4pc-burst", -1),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				if char.Energy > 0 {
					return nil, false
				}
				if char.StatusIsActive(buffstopkeyburst) {
					return nil, false
				}
				if atk.Info.AttackTag == attacks.AttackTagElementalBurst {
					return burstbuff, true
				}
				return nil, false
			},
		})
		c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
			atk := args[1].(*combat.AttackEvent)

			if atk.Info.ActorIndex != char.Index {
				return false
			}

			if atk.Info.AttackTag == attacks.AttackTagNormal {
				char.AddStatus(buffstopkeyburst, 6*60, true)
			}

			if atk.Info.AttackTag == attacks.AttackTagElementalBurst {
				char.AddStatus(buffstopkeyna, 6*60, true)
			}

			return false
		}, fmt.Sprintf("fodg-4pc-%v", char.Base.Key.String()))
	}

	return &s, nil
}
