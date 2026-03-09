package gildeddreams

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
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
	core.RegisterSetFunc(keys.GildedDreams, NewSet)
}

type Set struct {
	buff  []float64
	c     *core.Core
	char  *character.CharWrapper
	Index int
	Count int
}

func (s *Set) SetIndex(idx int) { s.Index = idx }
func (s *Set) GetCount() int    { return s.Count }
func (s *Set) Init() error {
	emCount := 0
	atkCount := 0

	for _, this := range s.c.Player.Chars() {
		if s.char.Index == this.Index {
			continue
		}
		if this.Base.Element != s.char.Base.Element {
			emCount++
		} else {
			atkCount++
		}
	}

	if emCount > 3 {
		emCount = 3
	}
	if atkCount > 3 {
		atkCount = 3
	}

	s.buff = make([]float64, attributes.EndStatType)
	s.buff[attributes.EM] = 50 * float64(emCount)
	s.buff[attributes.ATKP] = 0.14 * float64(atkCount)

	return nil
}

// 2セット効果: 元素熟知 +80
// 4セット効果: 元素反応発動後8秒間、装備キャラはパーティメンバーの元素タイプに基づくバフを獲得。
// 装備キャラと同じ元素タイプのパーティメンバー1人ごとに攻撃力14%増加、異なる元素タイプの
// パーティメンバー1人ごとに元素熟知50増加。上記バフはそれぞれ最大3人までカウント。
// この効果は8秒に1回発動可能。装備キャラがフィールド上にいなくても効果を発動可能。
func NewSet(c *core.Core, char *character.CharWrapper, count int, param map[string]int) (info.Set, error) {
	s := Set{
		c:     c,
		char:  char,
		Count: count,
	}

	if count >= 2 {
		m := make([]float64, attributes.EndStatType)
		m[attributes.EM] = 80
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("gd-2pc", -1),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}
	if count >= 4 {
		const icdKey = "gd-4pc-icd"
		add := func(args ...interface{}) bool {
			atk := args[1].(*combat.AttackEvent)
			if atk.Info.ActorIndex != char.Index {
				return false
			}
			if char.StatusIsActive(icdKey) {
				return false
			}
			char.AddStatus(icdKey, 8*60, true)

			char.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag("gd-4pc", 8*60),
				AffectedStat: attributes.NoStat,
				Amount: func() ([]float64, bool) {
					return s.buff, true
				},
			})
			c.Log.NewEvent("gilded dreams proc'd", glog.LogArtifactEvent, char.Index).
				Write("em", s.buff[attributes.EM]).
				Write("atk", s.buff[attributes.ATKP])
			return false
		}

		for i := event.ReactionEventStartDelim + 1; i < event.OnShatter; i++ {
			c.Events.Subscribe(i, add, fmt.Sprintf("gd-4pc-%v", char.Base.Key.String()))
		}
	}

	return &s, nil
}
