package martialartist

import (
	"fmt"

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
	core.RegisterSetFunc(keys.MartialArtist, NewSet)
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

	// 2セット: 通常攻撃と重撃のダメージが15%増加。
	if count >= 2 {
		m := make([]float64, attributes.EndStatType)
		m[attributes.DmgP] = 0.15
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("martialartist-2pc", -1),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				if atk.Info.AttackTag != attacks.AttackTagNormal && atk.Info.AttackTag != attacks.AttackTagExtra {
					return nil, false
				}
				return m, true
			},
		})
	}
	// 4セット: 元素スキル使用後、通常攻撃と重撃のダメージが8秒間25%増加。
	if count >= 4 {
		m := make([]float64, attributes.EndStatType)
		m[attributes.DmgP] = 0.25
		c.Events.Subscribe(event.OnSkill, func(args ...interface{}) bool {
			// 他のキャラクターのスキルでは発動しない
			if c.Player.Active() != char.Index {
				return false
			}
			// バフを追加
			char.AddAttackMod(character.AttackMod{
				Base: modifier.NewBaseWithHitlag("martialartist-4pc", 480), // 8s
				Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
					if atk.Info.AttackTag != attacks.AttackTagNormal && atk.Info.AttackTag != attacks.AttackTagExtra {
						return nil, false
					}
					return m, true
				},
			})
			return false
		}, fmt.Sprintf("martialartist-4pc-%v", char.Base.Key.String()))
	}

	return &s, nil
}
