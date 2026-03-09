package bloodstained

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterSetFunc(keys.BloodstainedChivalry, NewSet)
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

	// 2セット: 物理ダメージボーナス +25%
	if count >= 2 {
		m := make([]float64, attributes.EndStatType)
		m[attributes.PhyP] = 0.25
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("bloodstained-2pc", -1),
			AffectedStat: attributes.PhyP,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}

	// 4セット: 敵を撃破後、重撃ダメージ50%増加、スタミナ消費0になる（10秒間）。
	// 猪、リス、カエルなどの野生動物にも発動する。
	if count < 4 {
		return &s, nil
	}

	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = 0.50
	c.Events.Subscribe(event.OnTargetDied, func(args ...interface{}) bool {
		_, ok := args[0].(*enemy.Enemy)
		// 敵でなければ無視
		if !ok {
			return false
		}
		atk := args[1].(*combat.AttackEvent)
		// 別のキャラクターが敵を倒した場合は発動しない
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		// フィールド外では発動しない
		if c.Player.Active() != char.Index {
			return false
		}

		// 重撃ダメージ%部分
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBaseWithHitlag("bloodstained-4pc-dmg%", 600),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				if atk.Info.AttackTag != attacks.AttackTagExtra {
					return nil, false
				}
				return m, true
			},
		})

		// 重撃スタミナ部分
		// TODO: ヒットラグの影響を受けるべきか？（スタミナ割合Mod）
		c.Player.AddStamPercentMod("bloodstained-4pc-stamina", 600, func(a action.Action) (float64, bool) {
			if a == action.ActionCharge {
				return -1, false
			}
			return 0, false
		})

		return false
	}, fmt.Sprintf("bloodstained-4pc-%v", char.Base.Key.String()))

	return &s, nil
}
