package simulation

import (
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/stats"
)

type stateFn func(*Simulation) (stateFn, error)

func (s *Simulation) resFromCurrentState() stats.Result {
	return stats.Result{Seed: uint64(s.C.Seed), Duration: s.C.F + 1}
}

func (s *Simulation) run() (stats.Result, error) {
	// コアループの大まかな流れ:
	//  - 初期化:
	//		- セットアップ
	//		- フレームを1つ進める
	//		- キューフェーズへ移動
	//  - キューフェーズ:
	//		- 次のアクションを要求
	//		- 準備チェックフェーズへ移動
	//	- 準備チェックフェーズ
	//		- アクションが準備完了か確認（アニメーションとプレイヤーの両方）; 準備未完了ならフレームを進めて待つ
	//		- アクション実行フェーズへ移動
	//	- アクション実行フェーズ:
	//		- アクション前の待機がある場合; 待機が消化されるまでフレームを進める
	//		- アクションを実行しキューを空にする
	//		- 実行されたアクションがno-opの場合、直接キューフェーズへ移動
	//		- そうでなければ CanQueueAfter までフレームを進めてからキューフェーズへ移動
	//
	// フレーム進行は以下を実行する:
	//	- フレームカウンターを1増加
	//  - ティックを実行
	//  - エネルギー処理を確認
	//  - OnTick を発火
	//  - 終了チェックを実行
	//
	// 終了チェックの確認項目:
	//	- フレーム上限
	//  - 全敵が死亡
	//  - 残りアクションなし

	//TODO: ここでまだパニックをキャッチする必要があるか？ワーカー側で処理できないか？
	var err error
	for state := initialize; state != nil; {
		state, err = state(s)
		if err != nil {
			return s.resFromCurrentState(), err
		}
	}

	// err = s.eval.Exit()
	// if err != nil {
	// 	fmt.Println("evaluator already closed")
	// 	return s.resFromCurrentState(), err
	// }
	s.eval.Exit()

	err = s.eval.Err()
	if err != nil {
		return s.resFromCurrentState(), err
	}

	s.C.Events.Emit(event.OnSimEndedSuccessfully)

	return s.gatherResult(), nil
}

func (s *Simulation) gatherResult() stats.Result {
	res := stats.Result{
		Seed:        uint64(s.C.Seed),
		Duration:    s.C.F,
		TotalDamage: s.C.Combat.TotalDamage,
		DPS:         s.C.Combat.TotalDamage * 60 / float64(s.C.F),
		Characters:  make([]stats.CharacterResult, len(s.C.Player.Chars())),
		Enemies:     make([]stats.EnemyResult, s.C.Combat.EnemyCount()),
		EndStats:    make([]stats.EndStats, len(s.C.Player.Chars())),
	}

	for i := range s.cfg.Characters {
		res.Characters[i].Name = s.cfg.Characters[i].Base.Key.String()
	}

	for _, collector := range s.collectors {
		collector.Flush(s.C, &res)
	}

	return res
}

func (s *Simulation) popQueue() int {
	switch len(s.queue) {
	case 0:
	case 1:
		s.queue = s.queue[:0]
	default:
		s.queue = s.queue[1:]
	}
	return len(s.queue)
}

func initialize(s *Simulation) (stateFn, error) {
	go s.eval.Start()
	// durationが未設定の場合、90秒でシミュレーションを実行
	if s.cfg.Settings.Duration == 0 {
		// fmt.Println("no duration set, running for 90s")
		s.cfg.Settings.Duration = 90
	}
	s.C.Flags.DamageMode = s.cfg.Settings.DamageMode

	return s.advanceFrames(1, queuePhase)
}

func queuePhase(s *Simulation) (stateFn, error) {
	if s.noMoreActions {
		return s.advanceFrames(1, queuePhase)
	}
	s.eval.Continue()
	next, err := s.eval.NextAction()
	if err != nil {
		return nil, err
	}
	// evalにアクションが残っていない場合、1フレームスキップしてキューフェーズに戻る
	// 必要に応じて advanceFrame の終了処理に依存する
	if next == nil {
		s.noMoreActions = true
		// evalにアクションが残っていない場合と同じスキップ処理を行う
		return s.advanceFrames(1, queuePhase)
	}
	// sleepはフレームスキップしてから次をキューに戻すだけなのでここで処理する
	if next.Action == action.ActionWait {
		return s.handleWait(next)
	}
	// 次のアクションがdelayの場合、その後のアクションを即座にキューに追加できる
	if next.Action == action.ActionDelay {
		// 複数のdelayが連鎖する可能性があるためappendする
		delay := next.Param["f"]
		s.preActionDelay += delay
		s.C.Log.NewEvent(fmt.Sprintf("delay added %v, total: %v", delay, s.preActionDelay), glog.LogActionEvent, s.C.Player.Active()).
			Write("added", delay).
			Write("total", s.preActionDelay)
		return queuePhase, nil
	}
	// 重要: 次のキャラがアクティブでない場合、evaluatorが暗黙のスワップを追加する必要がある
	// evaluatorのエラーに備えてここでサニティチェックを追加する
	if next.Char != s.C.Player.ActiveChar().Base.Key && next.Action != action.ActionSwap {
		return nil, fmt.Errorf("internal error: requested next char %v is not active and next action is not swap", next.Char)
	}
	// TODO: キューを単一アイテムに変更することを検討。ここにスワップがないのでスライスは不要
	s.queue = append(s.queue, next)
	return actionReadyCheckPhase, nil
}

func actionReadyCheckPhase(s *Simulation) (stateFn, error) {
	//TODO: このサニティチェックはおそらく不要
	if len(s.queue) == 0 {
		return nil, errors.New("unexpected queue length is 0")
	}
	q := s.queue[0]

	// 次のキューアイテムが有効か確認する
	// 例: ほとんどの片手剣キャラは前のアクションが攻撃でない場合、重撃を実行できない
	char := s.C.Player.ActiveChar()
	if err := char.NextQueueItemIsValid(q.Char, q.Action, q.Param); err != nil {
		switch {
		case errors.Is(err, player.ErrInvalidChargeAction):
			return nil, fmt.Errorf("%v: %w", char.Base.Key, player.ErrInvalidChargeAction)
		default:
			return nil, err
		}
	}

	//TODO: このループは一度に1フレーム以上スキップするよう最適化すべき
	if err := s.C.Player.ReadyCheck(q.Action, q.Char, q.Param); err != nil {
		// アクションが準備完了になるまでこのフェーズを繰り返す
		switch {
		case errors.Is(err, player.ErrActionNotReady):
			s.C.Log.NewEvent(fmt.Sprintf("could not execute %v; action not ready", q.Action), glog.LogSimEvent, s.C.Player.Active())
			return s.advanceFrames(1, actionReadyCheckPhase)
		case errors.Is(err, player.ErrPlayerNotReady):
			return s.advanceFrames(1, actionReadyCheckPhase)
		case errors.Is(err, player.ErrActionNoOp):
			// ここでは何もしない
		default:
			return nil, err
		}
	}

	return executeActionPhase, nil
}

func (s *Simulation) handleWait(q *action.Eval) (stateFn, error) {
	// 既存の機能を維持するため、wait（sleepのエイリアス）は常に準備完了であり、
	// パラメータ f と同じ数だけ advanceFrames を呼び出す
	skip := q.Param["f"]
	// wait(0) であることが分かりやすいように別途ログに記録する
	if skip == 0 {
		s.C.Log.NewEvent("executed noop wait(0)", glog.LogActionEvent, s.C.Player.Active()).
			Write("f", skip)
	} else {
		s.C.Log.NewEvent("executed wait", glog.LogActionEvent, s.C.Player.Active()).
			Write("f", skip)
	}
	if l := s.popQueue(); l > 0 {
		// 既にキューにアクションがある場合はキューに戻らない
		return s.advanceFrames(skip, actionReadyCheckPhase)
	}
	return s.advanceFrames(skip, queuePhase)
}

func executeActionDelay(s *Simulation) (stateFn, error) {
	if s.preActionDelay > 0 {
		if !s.C.Player.ActiveChar().FramePausedOnHitlag() {
			s.preActionDelay--
		}
		return s.advanceFrames(1, executeActionDelay)
	}
	// delay後にアクションが利用不可になった場合に備えて準備チェックフェーズに戻る
	return actionReadyCheckPhase, nil
}

func executeActionPhase(s *Simulation) (stateFn, error) {
	//TODO: このサニティチェックはおそらく不要
	if len(s.queue) == 0 {
		return nil, errors.New("unexpected queue length is 0")
	}
	if s.preActionDelay > 0 {
		delay := s.preActionDelay
		s.C.Log.NewEvent(fmt.Sprintf("pre action delay: %v", delay), glog.LogActionEvent, s.C.Player.Active()).
			Write("delay", delay)
		return executeActionDelay, nil
	}
	q := s.queue[0]
	err := s.C.Player.Exec(q.Action, q.Char, q.Param)
	if err != nil {
		//TODO: このチェックはおそらく何もしない
		if errors.Is(err, player.ErrActionNoOp) {
			if l := s.popQueue(); l > 0 {
				// 既にキューにアクションがある場合はキューに戻らない
				return actionReadyCheckPhase, nil
			}
			return queuePhase, nil
		}
		// アクションは準備完了のはずなので、ここでのエラーは想定外
		// コンテキストを追加するためエラーをラップする
		return nil, fmt.Errorf("error encountered on %v executing %v: %w", q.Char.String(), q.Action.String(), err)
	}
	//TODO: ここのチェックはおそらく不要
	if l := s.popQueue(); l > 0 {
		// 既にキューにアクションがある場合はキューに戻らない
		return actionReadyCheckPhase, nil
	}

	return skipUntilCanQueue, nil
}

func skipUntilCanQueue(s *Simulation) (stateFn, error) {
	if !s.C.Player.CanQueueNextAction() {
		return s.advanceFrames(1, skipUntilCanQueue)
	}
	return queuePhase, nil
}

// advanceFrames はフレームを指定数だけ進め、各フレームで処理を実行する
func (s *Simulation) advanceFrames(f int, next stateFn) (stateFn, error) {
	for i := 0; i < f; i++ {
		done, err := s.nextFrame()
		if err != nil {
			return nil, err
		}
		if done {
			return nil, nil
		}
	}
	return next, nil
}

func (s *Simulation) nextFrame() (bool, error) {
	s.C.F++
	err := s.C.Tick()
	if err != nil {
		return false, err
	}
	s.handleEnergy()
	s.handleHurt()
	s.C.Events.Emit(event.OnTick)
	return s.stopCheck(), nil
}

func (s *Simulation) stopCheck() bool {
	if s.C.Combat.DamageMode {
		// アクションが残っていない場合は停止
		if s.noMoreActions {
			return true
		}
		// 全ターゲットが死亡を報告している場合は停止
		allDead := true
		for _, t := range s.C.Combat.Enemies() {
			if t.IsAlive() {
				allDead = false
				break
			}
		}
		return allDead
	}
	return s.C.F == int(s.cfg.Settings.Duration*60)
}

// TODO: 各関数が実際にerrorを返すようにしてdeferを削除する
//
//nolint:nonamedreturns,nakedret // 名前付き戻り値なしでは res, err の変更ができないため
func (s *Simulation) Run() (res stats.Result, err error) {
	defer func() {
		// パニックが発生した場合はリカバリする。それ以外の場合は err を nil に設定する。
		if r := recover(); r != nil {
			res = stats.Result{Seed: uint64(s.C.Seed), Duration: s.C.F + 1}
			err = fmt.Errorf("simulation panic occured: %v \n"+string(debug.Stack()), r)
		}
	}()
	res, err = s.run()
	return
}
