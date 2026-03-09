package instructor

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterSetFunc(keys.Instructor, NewSet)
}

type Set struct {
	Index int
	Count int
}

func (s *Set) SetIndex(idx int) { s.Index = idx }
func (s *Set) GetCount() int    { return s.Count }
func (s *Set) Init() error      { return nil }

// Instructor聖遺物セットの実装:
// 2セット効果: 元素熟知+80
// 4セット効果: 元素反応発動時、全パーティメンバーの元素熟知が8秒間120増加。
func NewSet(c *core.Core, char *character.CharWrapper, count int, param map[string]int) (info.Set, error) {
	s := Set{Count: count}

	if count >= 2 {
		m := make([]float64, attributes.EndStatType)
		m[attributes.EM] = 80
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("instructor-2pc", -1),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}
	if count >= 4 {
		m := make([]float64, attributes.EndStatType)
		m[attributes.EM] = 120

		// TODO: 複数の教官所持者で効果時間は延長されるか？
		add := func(args ...interface{}) bool {
			atk := args[1].(*combat.AttackEvent)
			// ボーナス発動にはキャラがフィールド上にいる必要がある
			if c.Player.Active() != char.Index {
				return false
			}
			// 元素反応の発生源は教官を装備したキャラクターでなければならない
			if atk.Info.ActorIndex != char.Index {
				return false
			}

			// 全キャラクターに元素熟知120を追加
			for _, this := range c.Player.Chars() {
				this.AddStatMod(character.StatMod{
					Base:         modifier.NewBaseWithHitlag("instructor-4pc", 480),
					AffectedStat: attributes.EM,
					Amount: func() ([]float64, bool) {
						return m, true
					},
				})
			}
			return false
		}

		for i := event.ReactionEventStartDelim + 1; i < event.OnShatter; i++ {
			c.Events.Subscribe(i, add, fmt.Sprintf("instructor-4pc-%v", char.Base.Key.String()))
		}
	}

	return &s, nil
}
