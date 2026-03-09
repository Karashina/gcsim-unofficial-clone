// package minazukiは通常攻撃アニメーション状態に基づいてトリガーされるアビリティの
// 共通実装を提供する（例: 行秋の元素爆発）
package minazuki

import (
	"errors"
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

// Watcher は状態変化を監視し、それに応じてトリガーする
type Watcher struct {
	// 必須データ
	key         keys.Char // ウォッチャーの名前。購読者のキーとして使用
	caster      *character.CharWrapper
	abil        string
	statusKey   string
	icdKey      string
	triggerFunc func()
	core        *core.Core
	tickerFreq  int

	// その他のフィールド（オプショナルのオーバーライド含む）
	state        action.AnimationState   // 監視対象の状態変化
	delayKey     model.AnimationDelayKey // ディレイ関数確認用のディレイキー
	shouldDelay  func() bool             // ディレイを適用すべきか判定する関数
	tickOnActive bool

	tickSrc int
}

type Config func(w *Watcher) error

func New(cfg ...Config) (*Watcher, error) {
	w := &Watcher{
		// デフォルト
		delayKey: model.InvalidAnimationDelayKey,
		state:    action.NormalAttackState,
	}
	for _, f := range cfg {
		err := f(w)
		if err != nil {
			return nil, err
		}
	}

	caster, ok := w.core.Player.ByKey(w.key)
	if !ok {
		return nil, errors.New("caster key is invalid")
	}
	w.caster = caster

	w.stateChangeHook()
	return w, nil
}

func WithMandatory(key keys.Char, abil, statusKey, icdKey string, tickerFreq int, triggerFunc func(), c *core.Core) Config {
	return func(w *Watcher) error {
		if abil == "" {
			return errors.New("ability name cannot be blank")
		}
		if statusKey == "" {
			return errors.New("status key cannot be blank")
		}
		if tickerFreq == 0 {
			return errors.New("ticker frequency cannot be 0")
		}
		w.key = key
		w.abil = abil
		w.statusKey = statusKey
		w.icdKey = icdKey
		w.triggerFunc = triggerFunc
		w.tickerFreq = tickerFreq
		w.core = c
		return nil
	}
}

func WithAnimationDelayCheck(key model.AnimationDelayKey, shouldDelay func() bool) Config {
	return func(w *Watcher) error {
		w.delayKey = key
		w.shouldDelay = shouldDelay
		return nil
	}
}

func WithTickOnActive(v bool) Config {
	return func(w *Watcher) error {
		w.tickOnActive = v
		return nil
	}
}

func WithAnimationState(s action.AnimationState) Config {
	return func(w *Watcher) error {
		w.state = s
		return nil
	}
}

func (w *Watcher) Kill() {
	w.tickSrc = -1
}

func (w *Watcher) stateChangeHook() {
	w.core.Events.Subscribe(event.OnStateChange, func(args ...interface{}) bool {
		next := args[1].(action.AnimationState)
		// 監視対象の状態でなければ無視
		if next != w.state {
			return false
		}

		// 遅延チェックが必要で、かつ遅延がある場合、この状態チェックの実行を遅延させる
		// ステータスがアクティブでない場合はこのチェックが不要なので
		// パフォーマンス的には効率が悪いが、可読性を優先する
		// パフォーマンスへの影響はそこまで大きくないはず
		if w.shouldDelay != nil { // TODO: 旧実装との互換性維持のため。削除すべき
			// if w.shouldDelay() {
			if delay := w.core.Player.ActiveChar().AnimationStartDelay(w.delayKey); delay > 0 {
				c := w.caster
				w.core.Log.NewEventBuildMsg(glog.LogDebugEvent, c.Index, w.abil, " delay on state change").
					Write("delay", delay)
				w.core.Tasks.Add(w.onStateChange(next), delay)
				return false
			}
		}
		w.onStateChange(next)()

		return false
	}, fmt.Sprintf("%v-burst-state-change-hook", w.key.String()))
}

func (w *Watcher) onStateChange(next action.AnimationState) func() {
	return func() {
		c := w.caster
		if !c.StatusIsActive(w.statusKey) {
			return
		}
		if w.icdKey != "" && c.StatusIsActive(w.icdKey) {
			w.core.Log.NewEventBuildMsg(glog.LogCharacterEvent, c.Index, w.abil, " not triggered on state change; on icd").
				Write("icd", c.StatusExpiry(w.icdKey)).
				Write("icd_key", w.icdKey)
			return
		}
		w.triggerFunc()
		w.core.Log.NewEventBuildMsg(glog.LogCharacterEvent, c.Index, w.abil, " triggered on state change").
			Write("state", next).
			Write("icd", c.StatusExpiry(w.icdKey)).
			Write("icd_key", w.icdKey)

		w.tickSrc = w.core.F
		w.queueTick(w.core.F)
	}
}

func (w *Watcher) queueTick(src int) {
	if w.tickerFreq <= 0 {
		return
	}

	c := w.caster
	if w.tickOnActive {
		c = w.core.Player.ActiveChar()
	}
	// ヒットラグ影響付きキューを使用
	c.QueueCharTask(w.tickerFunc(src), w.tickerFreq)
}

func (w *Watcher) tickerFunc(src int) func() {
	return func() {
		c := w.caster
		// バフが有効か確認
		if !c.StatusIsActive(w.statusKey) {
			w.core.Log.NewEventBuildMsg(glog.LogCharacterEvent, c.Index, w.abil, " not triggered on tick; on icd").
				Write("icd", c.StatusExpiry(w.icdKey)).
				Write("icd_key", w.icdKey)
			return
		}
		if w.tickSrc != src {
			w.core.Log.NewEventBuildMsg(glog.LogCharacterEvent, c.Index, w.abil, " tick check ignored, src diff").
				Write("src", src).
				Write("new src", w.tickSrc)
			return
		}
		// 正しいアニメーション状態でなくなったら停止
		state := w.core.Player.CurrentState()
		if state != w.state {
			w.core.Log.NewEventBuildMsg(glog.LogCharacterEvent, c.Index, w.abil, " tick check stopped, wrong state").
				Write("src", src).
				Write("state", state)
			return
		}
		if w.shouldDelay != nil && w.shouldDelay() {
			// 可読性のためifをネストしている...
			s := w.core.Player.CurrentStateStart()
			if w.core.F-s < w.core.Player.ActiveChar().AnimationStartDelay(w.delayKey) {
				w.core.Log.NewEventBuildMsg(glog.LogCharacterEvent, c.Index, w.abil, " not triggered; not enough time since normal state start").
					Write("current_state", state).
					Write("state_start", s)
				return
			}
		}
		// 遅延チェックがある場合、現在のフレームカウントが遅延を超えていることを確認
		w.core.Log.NewEventBuildMsg(glog.LogCharacterEvent, c.Index, w.abil, " triggered from ticker").
			Write("src", src).
			Write("state", state).
			Write("icd", c.StatusExpiry(w.statusKey))
		// 通常状態かつsrcが同じなので、ここでウェーブをトリガーできる
		w.triggerFunc()
		// 理論上はICDに引っかからないはず？
		// ヒットラグ影響付きキューを使用
		w.queueTick(src)
	}
}
