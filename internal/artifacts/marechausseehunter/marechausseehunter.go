package marechausseehunter

import (
	"fmt"
	"math"

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
	core.RegisterSetFunc(keys.MarechausseeHunter, NewSet)
}

type Set struct {
	stacks int
	core   *core.Core
	char   *character.CharWrapper
	buff   []float64
	Index  int
	Count  int
}

func (s *Set) SetIndex(idx int) { s.Index = idx }
func (s *Set) GetCount() int    { return s.Count }
func (s *Set) Init() error      { return nil }

func (s *Set) onChangeHP() {
	const buffKey = "mh-4pc"

	if !s.char.StatModIsActive(buffKey) {
		s.stacks = 0
	}
	if s.stacks < 3 {
		s.stacks++
	}

	s.char.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag(buffKey, 5*60),
		AffectedStat: attributes.CR,
		Amount: func() ([]float64, bool) {
			s.buff[attributes.CR] = 0.12 * float64(s.stacks)
			return s.buff, true
		},
	})
}

func NewSet(c *core.Core, char *character.CharWrapper, count int, param map[string]int) (info.Set, error) {
	s := Set{
		core:  c,
		char:  char,
		Count: count,
	}

	// 通常攻撃と重撃ダメージ+15%
	if count >= 2 {
		m := make([]float64, attributes.EndStatType)
		m[attributes.DmgP] = 0.15
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("mh-2pc", -1),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				if atk.Info.AttackTag != attacks.AttackTagNormal && atk.Info.AttackTag != attacks.AttackTagExtra {
					return nil, false
				}
				return m, true
			},
		})
	}

	// 現在のHPが増減すると、会心率12%増加5秒間。最大3スタック。
	if count < 4 {
		return &s, nil
	}

	s.buff = make([]float64, attributes.EndStatType)

	c.Events.Subscribe(event.OnPlayerHPDrain, func(args ...interface{}) bool {
		di := args[0].(*info.DrainInfo)
		if di.ActorIndex != char.Index {
			return false
		}
		if c.Player.Active() != char.Index {
			return false
		}
		if di.Amount <= 0 {
			return false
		}

		s.onChangeHP()
		return false
	}, fmt.Sprintf("mh-4pc-drain-%v", char.Base.Key.String()))

	c.Events.Subscribe(event.OnHeal, func(args ...interface{}) bool {
		index := args[1].(int)
		amount := args[2].(float64)
		overheal := args[3].(float64)
		if c.Player.Active() != char.Index {
			return false
		}
		if index != char.Index {
			return false
		}
		if amount <= 0 {
			return false
		}
		// 既にHP最大の場合は発動しない
		if math.Abs(amount-overheal) <= 1e-9 {
			return false
		}

		s.onChangeHP()
		return false
	}, fmt.Sprintf("mh-4pc-heal-%v", char.Base.Key.String()))

	// TODO: OnCharacterHurt？

	return &s, nil
}
