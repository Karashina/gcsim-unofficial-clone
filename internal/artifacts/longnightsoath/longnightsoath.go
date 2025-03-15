package longnightsoath

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
	"github.com/genshinsim/gcsim/pkg/core/stacks"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterSetFunc(keys.LongNightsOath, NewSet)
}

const (
	stackDuration    = 6 * 60
	buffICDKeyPlunge = "lo-4pc-icd-Plunge"
	buffICDKeyCharge = "lo-4pc-icd-Charge"
	buffICDKeySkill  = "lo-4pc-icd-Skill"
)

type Set struct {
	Index        int
	Count        int
	c            *core.Core
	char         *character.CharWrapper
	stackTracker *stacks.MultipleRefreshNoRemove
	buffStack    float64
}

func (s *Set) SetIndex(idx int) { s.Index = idx }
func (s *Set) GetCount() int    { return s.Count }
func (s *Set) Init() error      { return nil }

func NewSet(c *core.Core, char *character.CharWrapper, count int, param map[string]int) (info.Set, error) {
	s := Set{
		Count:        count,
		c:            c,
		char:         char,
		stackTracker: stacks.NewMultipleRefreshNoRemove(5, char.QueueCharTask, &c.F),
		buffStack:    0.15,
	}
	if count >= 2 {
		m := make([]float64, attributes.EndStatType)
		m[attributes.DmgP] = 0.25
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("lo-2pc", -1),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				if atk.Info.AttackTag == attacks.AttackTagPlunge {
					return m, true
				}
				return nil, false
			},
		})
	}
	if count >= 4 {
		m := make([]float64, attributes.EndStatType)
		c.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
			atk := args[1].(*combat.AttackEvent)

			if atk.Info.ActorIndex != char.Index {
				return false
			}

			if atk.Info.AttackTag == attacks.AttackTagPlunge && !char.StatusIsActive(buffICDKeyPlunge) {
				char.AddStatus(buffICDKeyPlunge, 60, true)
				s.stackTracker.Add(stackDuration)
				char.AddAttackMod(character.AttackMod{
					Base: modifier.NewBase("lo-4pc", stackDuration),
					Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
						switch atk.Info.AttackTag {
						case attacks.AttackTagPlunge:
						default:
							return nil, false
						}
						m[attributes.DmgP] = s.buffStack * float64(s.stackTracker.Count())
						return m, true
					},
				})
			}

			if atk.Info.AttackTag == attacks.AttackTagExtra && !char.StatusIsActive(buffICDKeyCharge) {
				char.AddStatus(buffICDKeyCharge, 60, true)
				s.stackTracker.Add(stackDuration)
				s.stackTracker.Add(stackDuration)
				char.AddAttackMod(character.AttackMod{
					Base: modifier.NewBase("lo-4pc", stackDuration),
					Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
						switch atk.Info.AttackTag {
						case attacks.AttackTagPlunge:
						default:
							return nil, false
						}
						m[attributes.DmgP] = s.buffStack * float64(s.stackTracker.Count())
						return m, true
					},
				})
			}

			if !char.StatusIsActive(buffICDKeySkill) {
				if atk.Info.AttackTag == attacks.AttackTagElementalArt || atk.Info.AttackTag == attacks.AttackTagElementalArtHold {
					char.AddStatus(buffICDKeySkill, 60, true)
					s.stackTracker.Add(stackDuration)
					s.stackTracker.Add(stackDuration)
					char.AddAttackMod(character.AttackMod{
						Base: modifier.NewBase("lo-4pc", stackDuration),
						Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
							switch atk.Info.AttackTag {
							case attacks.AttackTagPlunge:
							default:
								return nil, false
							}
							m[attributes.DmgP] = s.buffStack * float64(s.stackTracker.Count())
							return m, true
						},
					})
				}
			}

			return false
		}, fmt.Sprintf("lo-4pc-%v", char.Base.Key.String()))
	}
	return &s, nil
}
