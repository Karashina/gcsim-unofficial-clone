package geo

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/construct"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 1凸:
// 岩潮の範囲内のパーティメンバーは会心率10%増加、
//
//	中断耐性が向上。
func (c *Traveler) c1(ticks int) func() {
	return func() {
		// 6凸では異なるQフィールドが共存可能:
		// - QのCT復帰後(15秒後)でQフィールド消滅前(20秒前)に再度Qを発動
		// そのため両方がc1ティックをキューに入れても問題ない

		// Qの構造物がない場合、バフを適用せず次のティックもキューしない
		if c.Core.Constructs.CountByType(construct.GeoConstructTravellerBurst) == 0 {
			return
		}

		// 各Qフィールドが6凸未満/6凸でそれぞれ15/20回だけティックするようにする
		if ticks > c.c1TickCount {
			return
		}

		c.Core.Log.NewEvent("geo-traveler field ticking", glog.LogCharacterEvent, -1).
			Write("tick_number", ticks)

		// アクティブキャラに1凸バフを2秒間適用
		if c.Core.Combat.Player().IsWithinArea(c.burstArea) {
			m := make([]float64, attributes.EndStatType)
			m[attributes.CR] = .1

			active := c.Core.Player.ActiveChar()
			active.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag("geo-traveler-c1", 120), // 2s
				AffectedStat: attributes.CR,
				Amount: func() ([]float64, bool) {
					return m, true
				},
			})
		}

		// 1秒後に再チェック
		ticks += 1
		c.Core.Tasks.Add(c.c1(ticks), 60)
	}
}
