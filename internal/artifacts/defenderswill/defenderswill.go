package defenderswill

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterSetFunc(keys.DefendersWill, NewSet)
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

	// 2セット: 防御力 +30%
	if count >= 2 {
		m := make([]float64, attributes.EndStatType)
		m[attributes.DEFP] = 0.30
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("defenderswill-2pc", -1),
			AffectedStat: attributes.DEFP,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}
	// TODO: プレイヤーダメージが正確でないため現時点では実装不要
	// 4セット: パーティ内の異なる元素ごとに、装備者の対応する元素耐性が30%増加。
	if count >= 4 {
		c.Log.NewEvent("defenderswill-4pc not implemented", glog.LogArtifactEvent, char.Index)
	}

	return &s, nil
}
