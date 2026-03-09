package action

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/stats"
)

func init() {
	stats.Register(stats.Config{
		Name: "action",
		New:  NewStat,
	})
}

type buffer struct {
	energySpent    []float64
	failures       [][]stats.ActionFailInterval
	activeFailures []map[action.Action]activeFailure
	actionEvents   [][]stats.ActionEvent
}

type activeFailure struct {
	start  int
	reason action.Failure
}

func (b buffer) addFailure(core *core.Core, char int, active activeFailure) {
	interval := stats.ActionFailInterval{
		Start:  active.start,
		End:    core.F,
		Reason: active.reason.String(),
	}

	// TODO: 間隔を最低x（5？）に制限
	b.failures[char] = append(b.failures[char], interval)
}

func NewStat(core *core.Core) (stats.Collector, error) {
	out := buffer{
		energySpent:    make([]float64, len(core.Player.Chars())),
		failures:       make([][]stats.ActionFailInterval, len(core.Player.Chars())),
		activeFailures: make([]map[action.Action]activeFailure, len(core.Player.Chars())),
		actionEvents:   make([][]stats.ActionEvent, len(core.Player.Chars())),
	}

	for i := 0; i < len(out.activeFailures); i++ {
		out.activeFailures[i] = make(map[action.Action]activeFailure)
	}

	core.Events.Subscribe(event.OnActionExec, func(args ...interface{}) bool {
		char := args[0].(int)
		e := args[1].(action.Action)

		if e == action.ActionBurst {
			out.energySpent[char] += core.Player.Chars()[char].EnergyMax
		}

		// TODO: ActionIdの設定
		event := stats.ActionEvent{
			Frame:  core.F,
			Action: e.String(),
		}
		out.actionEvents[char] = append(out.actionEvents[char], event)

		if active, ok := out.activeFailures[char][e]; ok {
			out.addFailure(core, char, active)
			delete(out.activeFailures[char], e)
		}
		return false
	}, "stats-action-exec-log")

	core.Events.Subscribe(event.OnActionFailed, func(args ...interface{}) bool {
		char := args[0].(int)
		e := args[1].(action.Action)
		reason := args[3].(action.Failure)

		// アクションが成功するまで試行を続ける前提。
		// アクションの試行を中断した場合、これは正確ではなくなる
		// TODO: このエッジケースに対応するためアクションIDで追跡する？
		if _, ok := out.activeFailures[char][e]; !ok {
			out.activeFailures[char][e] = activeFailure{
				start:  core.F,
				reason: reason,
			}
		}

		return false
	}, "stats-action-failed-log")

	return &out, nil
}

func (b buffer) Flush(core *core.Core, result *stats.Result) {
	for c := 0; c < len(core.Player.Chars()); c++ {
		for _, active := range b.activeFailures[c] {
			b.addFailure(core, c, active)
		}

		result.Characters[c].FailedActions = b.failures[c]
		result.Characters[c].EnergySpent = b.energySpent[c]
		result.Characters[c].ActionEvents = b.actionEvents[c]
	}
}
