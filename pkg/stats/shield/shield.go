package shield

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/shield"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/stats"
)

var elements = [...]attributes.Element{
	attributes.Anemo,
	attributes.Cryo,
	attributes.Electro,
	attributes.Geo,
	attributes.Hydro,
	attributes.Pyro,
	attributes.Dendro,
	attributes.Physical,
}

const normalized = "normalized"

func init() {
	stats.Register(stats.Config{
		Name: "shield",
		New:  NewStat,
	})
}

type buffer struct {
	shields map[string][]stats.ShieldInterval
}

func NewStat(core *core.Core) (stats.Collector, error) {
	out := buffer{
		shields: make(map[string][]stats.ShieldInterval),
	}

	core.Events.Subscribe(event.OnShielded, func(args ...interface{}) bool {
		shield := args[0].(shield.Shield)
		name := shield.Desc()
		bonus := core.Player.Shields.ShieldBonus()

		interval := stats.ShieldInterval{
			Start: core.F,
			End:   shield.Expiry(),
			HP:    make(map[string]float64),
		}

		var normalizedHP float64
		for _, e := range elements {
			hp := shield.ShieldStrength(e, bonus)
			interval.HP[e.String()] = hp
			normalizedHP += hp
		}
		interval.HP[normalized] = normalizedHP / float64(len(elements))

		// このシールド種別の最初のインスタンス
		if _, ok := out.shields[name]; !ok {
			out.shields[name] = make([]stats.ShieldInterval, 0)
			out.shields[name] = append(out.shields[name], interval)
			return false
		}

		prevIndex := len(out.shields[name]) - 1
		prevInterval := out.shields[name][prevIndex]

		// 前のシールドが期限切れ前に更新された
		if prevInterval.End >= interval.Start {
			// HPステータスが同じ場合、区間をマージ
			if same(prevInterval.HP, interval.HP) {
				// TODO: max はここでは不要？
				prevInterval.End = max(prevInterval.End, interval.End)
				out.shields[name][prevIndex] = prevInterval
				return false
			}

			// 値が異なる、前の区間を早期終了
			prevInterval.End = interval.Start
			out.shields[name][prevIndex] = prevInterval
		}

		out.shields[name] = append(out.shields[name], interval)
		return false
	}, "stats-shield-log")

	// TODO: ターゲット指定イベントで置き換えるべき（シールドステータス変更 + キャラ交代時）
	core.Events.Subscribe(event.OnTick, func(args ...interface{}) bool {
		bonus := core.Player.Shields.ShieldBonus()

		for _, shield := range core.Player.Shields.List() {
			interval := stats.ShieldInterval{
				Start: core.F,
				End:   shield.Expiry(),
				HP:    make(map[string]float64),
			}

			var normalizedHP float64
			for _, e := range elements {
				hp := shield.ShieldStrength(e, bonus)
				interval.HP[e.String()] = hp
				normalizedHP += hp
			}
			interval.HP[normalized] = normalizedHP / float64(len(elements))

			prevIndex := len(out.shields[shield.Desc()]) - 1
			prevInterval := out.shields[shield.Desc()][prevIndex]
			if !same(prevInterval.HP, interval.HP) {
				if prevInterval.Start == interval.Start {
					// シールドが最初のフレームで再計算される特殊ケース
					out.shields[shield.Desc()][prevIndex] = interval
				} else {
					prevInterval.End = interval.Start
					out.shields[shield.Desc()][prevIndex] = prevInterval
					out.shields[shield.Desc()] = append(out.shields[shield.Desc()], interval)
				}
			}
		}
		return false
	}, "stats-shield-tick-log")

	return &out, nil
}

func (b buffer) Flush(core *core.Core, result *stats.Result) {
	shields := make([]stats.ShieldStats, 0, len(b.shields))
	for name, sb := range b.shields {
		shield := stats.ShieldStats{
			Name:      name,
			Intervals: sb,
		}
		shields = append(shields, shield)
	}

	result.ShieldResults = stats.ShieldResult{
		Shields:         shields,
		EffectiveShield: computeEffective(b.shields),
	}
}

func same(i, j map[string]float64) bool {
	for _, e := range elements {
		if i[e.String()] != j[e.String()] {
			return false
		}
	}
	return true
}
