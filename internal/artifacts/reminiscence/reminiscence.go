package reminiscence

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
	core.RegisterSetFunc(keys.ShimenawasReminiscence, NewSet)
}

type Set struct {
	cd    int
	Index int
	Count int
}

func (s *Set) SetIndex(idx int) { s.Index = idx }
func (s *Set) GetCount() int    { return s.Count }
func (s *Set) Init() error      { return nil }

func NewSet(c *core.Core, char *character.CharWrapper, count int, param map[string]int) (info.Set, error) {
	s := Set{Count: count}
	s.cd = -1

	if count >= 2 {
		m := make([]float64, attributes.EndStatType)
		m[attributes.ATKP] = 0.18
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("shim-2pc", -1),
			AffectedStat: attributes.ATKP,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}
	// 11:51 AM] Episodde｜ShimenawaChildePeddler: Basically I found out that the fox set energy tax have around a 10 frame delay.
	// 元素スキル使用後の10フレーム以内に元素爆発を使用してエネルギー消費15を回避できるかテスト。タルタリヤで可能
	// この発見は #energy-drain-effects-have-a-delay に記載されている
	if count >= 4 {
		const icdKey = "shim-4pc-icd"
		icd := 600 // 10s * 60

		m := make([]float64, attributes.EndStatType)
		m[attributes.DmgP] = 0.50
		c.Events.Subscribe(event.OnSkill, func(args ...interface{}) bool {
			if c.Player.Active() != char.Index {
				return false
			}
			if char.Energy < 15 {
				return false
			}
			if char.StatusIsActive(icdKey) {
				return false
			}
			char.AddStatus(icdKey, icd, true)

			// エネルギーを15消費、通常攻撃/重撃/落下攻撃ダメージを50%増加
			c.Tasks.Add(func() {
				char.AddEnergy("shim-4pc", -15)
			}, 10)

			char.AddAttackMod(character.AttackMod{
				Base: modifier.NewBaseWithHitlag("shim-4pc", 60*10),
				Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
					switch atk.Info.AttackTag {
					case attacks.AttackTagNormal:
					case attacks.AttackTagExtra:
					case attacks.AttackTagPlunge:
					default:
						return nil, false
					}
					return m, true
				},
			})

			return false
		}, fmt.Sprintf("shim-4pc-%v", char.Base.Key.String()))
	}

	return &s, nil
}
