package exile

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterSetFunc(keys.TheExile, NewSet)
}

type Set struct {
	Index int
	Count int
}

func (s *Set) SetIndex(idx int) { s.Index = idx }
func (s *Set) GetCount() int    { return s.Count }
func (s *Set) Init() error      { return nil }

// 2セット効果: 元素チャージ効率 +20%
// 4セット効果: 元素爆発使用時、装備者以外の全パーティメンバーに2秒ごとにエネルギーを2回復（6秒間）。この効果は重複しない。
func NewSet(c *core.Core, char *character.CharWrapper, count int, param map[string]int) (info.Set, error) {
	s := Set{Count: count}

	if count >= 2 {
		m := make([]float64, attributes.EndStatType)
		m[attributes.ER] = 0.20
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("exile-2pc", -1),
			AffectedStat: attributes.ER,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}

	const buffKey = "exile-4pc"
	buffDuration := 360 // 6s * 60

	if count >= 4 {
		c.Events.Subscribe(event.OnBurst, func(args ...interface{}) bool {
			if c.Player.Active() != char.Index {
				return false
			}

			// TODO: 複数の旧貴族所持者で効果時間は延長されるか？
			// for now: if exile is still ticking on at least one char then reject new exile buff
			for _, x := range c.Player.Chars() {
				this := x
				if this.StatusIsActive(buffKey) {
					return false
				}
			}

			for _, x := range c.Player.Chars() {
				this := x
				if char.Index == this.Index {
					continue
				}
				// 装備者以外の全パーティメンバーに亡命ステータスを追加
				this.AddStatus(buffKey, buffDuration, true)
				// 3ティック
				for i := 120; i <= 360; i += 120 {
					// 亡命のティックはヒットラグの影響を受ける
					this.QueueCharTask(func() {
						this.AddEnergy("exile-4pc", 2)
					}, i)
				}
			}

			return false
		}, fmt.Sprintf("exile-4pc-%v", char.Base.Key.String()))
	}

	return &s, nil
}
