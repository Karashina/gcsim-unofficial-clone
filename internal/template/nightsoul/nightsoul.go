package nightsoul

import (
	"fmt"
	"slices"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
)

const NightsoulBlessingStatus = "nightsoul-blessing"
const delayEventKey = "ns-extend-state"

type State struct {
	char            *character.CharWrapper
	c               *core.Core
	nightsoulPoints float64
	ExitStateF      int
	MaxPoints       float64
	extendNsStates  []action.AnimationState
}

func New(c *core.Core, char *character.CharWrapper) *State {
	t := &State{
		char:       char,
		c:          c,
		ExitStateF: -1,
		MaxPoints:  -1.0, // no limits
	}
	return t
}

// NSの有効期限切れを防止するアニメーション状態のセットを設定する
// ここで設定した状態中にNSが期限切れになる場合、状態の終了まで遅延させる
func (s *State) SetExtendNsStates(states []action.AnimationState) {
	s.extendNsStates = states
}

func (s *State) Duration() int {
	return s.char.StatusDuration(NightsoulBlessingStatus)
}

// 該当する場合NSステータスの現在の持続時間を変更し、期限切れ時にコールバックを適用する。
// NSが現在アクティブでない場合は何もしない。
// ExitNightsoul()が使用された場合、コールバックは呼ばれない
func (s *State) SetNightsoulExitTimer(duration int, cb func()) {
	if !s.HasBlessing() {
		return
	}

	if s.Duration() != duration {
		s.char.AddStatus(NightsoulBlessingStatus, duration, true)
	}

	src := s.c.F + duration
	s.ExitStateF = src
	s.char.QueueCharTask(func() {
		if s.ExitStateF != src {
			return
		}

		if !slices.Contains(s.extendNsStates, s.c.Player.CurrentState()) {
			cb()
			return
		}

		// キャラがフィールド外の場合、アニメーションで状態を延長できない
		if !s.c.Player.CharIsActive(s.char.Base.Key) {
			cb()
			return
		}

		// プレイヤーが状態中のためNSが期限切れすべきでない場合、状態終了まで期限を遅延する
		evtKey := fmt.Sprintf("%v-%v", delayEventKey, s.char.Base.Key.String())
		f := func(...interface{}) bool {
			if s.ExitStateF == src {
				cb()
			}
			return true
		}
		s.c.Events.Subscribe(event.OnStateChange, f, evtKey)

		// 状態終了時に削除されるまでNSを延長
		s.char.AddStatus(NightsoulBlessingStatus, -1, false)
		s.c.Log.NewEvent("Action extending timed Nightsoul Blessing",
			glog.LogActionEvent,
			s.char.Index)
	}, duration)
}

// 指定されたポイントでNS祝福に入る。
// 持続時間が無限でない場合、持続時間経過時にNSを終了し、オプションでCBをトリガーする。
// 有効期限前に持続時間が変更された場合、CBは呼ばれない。
func (s *State) EnterTimedBlessing(amount float64, duration int, cb func()) {
	s.nightsoulPoints = amount
	s.char.AddStatus(NightsoulBlessingStatus, duration, true)

	if cb == nil {
		cb = s.ExitBlessing
	}
	if duration > 0 {
		s.SetNightsoulExitTimer(duration, cb)
	}
	s.c.Log.NewEvent("enter nightsoul blessing", glog.LogCharacterEvent, s.char.Index).
		Write("points", s.nightsoulPoints).
		Write("duration", duration)
}

func (s *State) EnterBlessing(amount float64) {
	s.EnterTimedBlessing(amount, -1, nil)
}

func (s *State) ExitBlessing() {
	s.ExitStateF = -1
	s.char.DeleteStatus(NightsoulBlessingStatus)
	s.c.Log.NewEvent("exit nightsoul blessing", glog.LogCharacterEvent, s.char.Index)
}

func (s *State) HasBlessing() bool {
	return s.char.StatusIsActive(NightsoulBlessingStatus)
}

func (s *State) GeneratePoints(amount float64) {
	prevPoints := s.nightsoulPoints
	s.nightsoulPoints += amount
	s.clampPoints()
	s.c.Events.Emit(event.OnNightsoulGenerate, s.char.Index, amount)
	s.c.Log.NewEvent("generate nightsoul points", glog.LogCharacterEvent, s.char.Index).
		Write("previous points", prevPoints).
		Write("amount", amount).
		Write("final", s.nightsoulPoints)
}

func (s *State) ConsumePoints(amount float64) {
	prevPoints := s.nightsoulPoints
	s.nightsoulPoints -= amount
	s.clampPoints()
	s.c.Events.Emit(event.OnNightsoulConsume, s.char.Index, amount)
	s.c.Log.NewEvent("consume nightsoul points", glog.LogCharacterEvent, s.char.Index).
		Write("previous points", prevPoints).
		Write("amount", amount).
		Write("final", s.nightsoulPoints)
}

func (s *State) ClearPoints() {
	amt := s.nightsoulPoints
	s.nightsoulPoints = 0
	s.c.Events.Emit(event.OnNightsoulConsume, s.char.Index, amt)
	s.c.Log.NewEvent("clear nightsoul points", glog.LogCharacterEvent, s.char.Index).
		Write("previous points", amt)
}

func (s *State) clampPoints() {
	if s.MaxPoints > 0 && s.nightsoulPoints > s.MaxPoints {
		s.nightsoulPoints = s.MaxPoints
	} else if s.nightsoulPoints < 0 {
		s.nightsoulPoints = 0
	}
}

func (s *State) Points() float64 {
	return s.nightsoulPoints
}

func (s *State) Condition(fields []string) (any, error) {
	switch fields[1] {
	case "state":
		return s.HasBlessing(), nil
	case "points":
		return s.Points(), nil
	case "duration":
		return s.Duration(), nil
	default:
		return nil, fmt.Errorf("invalid nightsoul condition: %v", fields[1])
	}
}
