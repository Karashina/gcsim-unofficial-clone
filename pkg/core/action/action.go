// Package action はキャラクターが実行可能な有効なアクションを定義する
package action

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
)

// TODO: メモリ割り当て削減のため sync.Pool を追加する
type Info struct {
	Frames              func(next Action) int `json:"-"`
	AnimationLength     int
	CanQueueAfter       int
	State               AnimationState
	FramePausedOnHitlag func() bool               `json:"-"`
	OnRemoved           func(next AnimationState) `json:"-"`
	// 以下はログを正しく出力するためにのみ公開
	TimePassed           float64
	NormalizedTimePassed float64
	UseNormalizedTime    func(next Action) bool
	// 非公開フィールド
	queued []queuedAction
}

// Eval はシミュレーションアクションを表す
type Eval struct {
	Char   keys.Char
	Action Action
	Param  map[string]int
}

// Evaluator は次のアクションを取得するメソッドを提供する
type Evaluator interface {
	NextAction() (*Eval, error) // NextAction should reuturn the next action, or nil if no actions left
	Continue()
	Exit() error
	Err() error
	Start()
}

type queuedAction struct {
	f     func()
	delay float64
}

func (a *Info) QueueAction(f func(), delay int) {
	a.queued = append(a.queued, queuedAction{f: f, delay: float64(delay)})
}

func (a *Info) CanQueueNext() bool {
	return a.TimePassed >= float64(a.CanQueueAfter)
}

func (a *Info) CanUse(next Action) bool {
	if a.UseNormalizedTime != nil && a.UseNormalizedTime(next) {
		return a.NormalizedTimePassed >= float64(a.Frames(next))
	}
	// 凍結中は何も使用できない
	if a.FramePausedOnHitlag != nil && a.FramePausedOnHitlag() {
		return false
	}
	return a.TimePassed >= float64(a.Frames(next))
}

func (a *Info) AnimationState() AnimationState {
	return a.State
}

func (a *Info) Tick() bool {
	a.NormalizedTimePassed++ // これは常にインクリメント
	// 時間が進むのはヒットラグ関数がないか、一時停止していない場合のみ
	if a.FramePausedOnHitlag == nil || !a.FramePausedOnHitlag() {
		a.TimePassed++
	}

	// timePassed > delay の全アクションを実行し、スライスから削除
	//
	if a.queued != nil {
		n := 0
		for i := 0; i < len(a.queued); i++ {
			if a.queued[i].delay <= a.TimePassed {
				a.queued[i].f()
			} else {
				a.queued[n] = a.queued[i]
				n++
			}
		}
		a.queued = a.queued[:n]
	}

	// アニメーションが終了したかチェック
	if a.TimePassed > float64(a.AnimationLength) {
		// 削除処理
		if a.OnRemoved != nil {
			a.OnRemoved(Idle)
		}
		return true
	}

	return false
}

type Action int

const (
	InvalidAction Action = iota
	ActionSkill
	ActionBurst
	ActionAttack
	ActionCharge
	ActionHighPlunge
	ActionLowPlunge
	ActionAim
	ActionDash
	ActionJump
	// 以下のアクションは実装がない
	ActionSwap
	ActionWalk
	ActionWait  // キャラクターが立ち止まって待機する
	ActionDelay // 次のアクション実行前の遅延
	EndActionType
	// 以下はフレーム目的のみで使用するため end の後に配置
	ActionSkillHoldFramesOnly
)

var astr = []string{
	"invalid",
	"skill",
	"burst",
	"attack",
	"charge",
	"high_plunge",
	"low_plunge",
	"aim",
	"dash",
	"jump",
	"swap",
	"walk",
	"wait",
	"delay",
}

func (a Action) String() string {
	return astr[a]
}

func (a Action) MarshalJSON() ([]byte, error) {
	return json.Marshal(astr[a])
}

func (a *Action) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	s = strings.ToLower(s)
	for i, v := range astr {
		if v == s {
			*a = Action(i)
			return nil
		}
	}
	return errors.New("unrecognized action")
}

func StringToAction(s string) Action {
	for i, v := range astr {
		if v == s {
			return Action(i)
		}
	}
	return InvalidAction
}

type AnimationState int

const (
	Idle AnimationState = iota
	NormalAttackState
	ChargeAttackState
	PlungeAttackState
	SkillState
	BurstState
	AimState
	DashState
	JumpState
	WalkState
	SwapState
)

var statestr = []string{
	"idle",
	"normal",
	"charge",
	"plunge",
	"skill",
	"burst",
	"aim",
	"dash",
	"jump",
	"walk",
	"swap",
}

func (a AnimationState) String() string {
	return statestr[a]
}
