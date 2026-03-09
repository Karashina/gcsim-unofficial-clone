// animation パッケージは、任意のフレームにおける現在のアニメーション状態と、
// 現在のフレームがアニメーションロック中かどうかを追跡するシンプルな方法を提供する
package animation

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/task"
)

type AnimationHandler struct { //nolint:revive // cannot just name this Handler because then there is a conflict with Handler in player package
	f      *int
	events event.Eventter
	log    glog.Logger
	tasks  task.Tasker

	char    int
	started int
	lastAct action.Action
	aniEvt  *action.Info

	state       action.AnimationState
	stateExpiry int

	debug bool
	event glog.Event
}

func New(f *int, debug bool, log glog.Logger, events event.Eventter, tasks task.Tasker) *AnimationHandler {
	h := &AnimationHandler{
		f:      f,
		log:    log,
		events: events,
		tasks:  tasks,
		debug:  debug,
	}
	return h
}

// IsAnimationLocked は現在のフレームで次のアクションを実行できない場合にtrueを返す
func (h *AnimationHandler) IsAnimationLocked(next action.Action) bool {
	if h.aniEvt == nil {
		return false
	}
	// アクションはtick処理後、次のフレームに進む直前（フレーム末尾）で実行される
	//
	// あるアクションのアニメーションが20フレーム必要な場合、
	// f >= s + 20 のとき準備完了となる
	//
	// つまり現在のフレームを含めて20フレーム継続したことになる
	// fmt.Printf("animation check; current frame %v, animation duration %v\n", *h.f, h.info.Frames(next))
	return !h.aniEvt.CanUse(next)
}

// CanQueueNextAction は現在のフレームで次のアクションのキューイングを開始できる場合にtrueを返す
func (h *AnimationHandler) CanQueueNextAction() bool {
	if h.aniEvt == nil {
		return true
	}
	return h.aniEvt.CanQueueNext()
}

func (h *AnimationHandler) SetActionUsed(char int, act action.Action, evt *action.Info) {
	// まだアクティブな場合は前のアニメーションを削除
	if h.aniEvt != nil {
		if h.aniEvt.OnRemoved != nil {
			h.aniEvt.OnRemoved(evt.State)
		}
		h.logEnded()
	}
	// 次のアニメーションを設定
	h.char = char
	h.started = *h.f
	h.aniEvt = evt
	h.events.Emit(event.OnStateChange, h.state, evt.State)
	h.state = evt.State
	h.stateExpiry = *h.f + evt.AnimationLength
	h.lastAct = act
	if h.debug {
		h.event = h.log.NewEvent(fmt.Sprintf("%v started", act.String()), glog.LogHitlagEvent, char).
			Write("AnimationLength", evt.AnimationLength).
			Write("CanQueueAfter", evt.CanQueueAfter).
			Write("State", evt.State.String())
		for i := action.Action(0); i < action.EndActionType; i++ {
			h.event.Write(i.String(), evt.Frames(i))
		}
	}
}

func (h *AnimationHandler) CurrentState() action.AnimationState {
	if h.aniEvt == nil {
		return action.Idle
	}
	return h.state
}

func (h *AnimationHandler) CurrentStateStart() int {
	return h.started
}

func (h *AnimationHandler) Tick() {
	if h.aniEvt != nil && h.aniEvt.Tick() {
		h.logEnded()
		h.events.Emit(event.OnStateChange, h.state, action.Idle)
		h.state = action.Idle
		h.aniEvt = nil
	}
}

func (h *AnimationHandler) logEnded() {
	if !h.debug {
		return
	}
	h.event.SetEnded(*h.f)
	h.log.NewEvent(
		fmt.Sprintf("%v from %v ended, time passed: %v (actual: %v)", h.lastAct, h.started, h.aniEvt.TimePassed, h.aniEvt.NormalizedTimePassed),
		glog.LogHitlagEvent,
		h.char,
	)
}
