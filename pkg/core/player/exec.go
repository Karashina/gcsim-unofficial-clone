package player

import (
	"errors"
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
)

// ErrActionNotReady は要求されたアクションがまだ準備できていない場合に返される。
// 以下の理由が考えられる:
//   - エネルギー不足（元素爆発のみ）
//   - スキルがクールダウン中
//   - プレイヤーがアニメーション中
var (
	// exec固有のエラー
	ErrActionNotReady        = errors.New("action is not ready yet; cannot be executed")
	ErrPlayerNotReady        = errors.New("player still in animation; cannot execute action")
	ErrInvalidAirborneAction = errors.New("player must use low_plunge or high_plunge while airborne")
	ErrActionNoOp            = errors.New("action is a noop")
	// キャラクター間で共有されるエラー
	ErrInvalidChargeAction = errors.New("need to use attack right before charge")
)

// ReadyCheck はアクションが実行可能ならnilを返し、そうでなければ理由を表すエラーを返す
func (h *Handler) ReadyCheck(t action.Action, k keys.Char, param map[string]int) error {
	// アニメーション状態をチェック
	if h.IsAnimationLocked(t) {
		return ErrPlayerNotReady
	}
	char := h.chars[h.active]
	// エネルギー、クールダウン等をチェック
	//TODO: キャラクター実装で重撃/ダッシュのスタミナデフォルトチェックがあることを確認する
	// 綾華/モナのドレインと通常消費の違いに対応する必要がある
	if ok, reason := char.ActionReady(t, param); !ok {
		h.Events.Emit(event.OnActionFailed, h.active, t, param, reason)
		return ErrActionNotReady
	}

	stamCheck := func(t action.Action, param map[string]int) (float64, bool) {
		req := h.AbilStamCost(char.Index, t, param)
		return req, h.Stam >= req
	}

	switch t {
	case action.ActionCharge: // require special calc for stam
		amt, ok := stamCheck(t, param)
		if !ok {
			h.Log.NewEvent("insufficient stam: charge attack", glog.LogWarnings, -1).
				Write("have", h.Stam).
				Write("cost", amt)
			h.Events.Emit(event.OnActionFailed, h.active, t, param, action.InsufficientStamina)
			return ErrActionNotReady
		}
	case action.ActionDash: // スタミナの特殊計算が必要
		// ダッシュはアクション自体で処理される
		amt, ok := stamCheck(t, param)
		if !ok {
			h.Log.NewEvent("insufficient stam: dash", glog.LogWarnings, -1).
				Write("have", h.Stam).
				Write("cost", amt)
			h.Events.Emit(event.OnActionFailed, h.active, t, param, action.InsufficientStamina)
			return ErrActionNotReady
		}

		// ダッシュがまだクールダウン中でロックされている場合、CDが切れるまで再ダッシュ不可
		if h.DashLockout && h.DashCDExpirationFrame > *h.F {
			h.Log.NewEvent("dash on cooldown", glog.LogWarnings, -1).
				Write("dash_cd_expiration", h.DashCDExpirationFrame-*h.F)
			h.Events.Emit(event.OnActionFailed, h.active, t, param, action.DashCD)
			return ErrActionNotReady
		}
	case action.ActionSwap:
		if h.active == h.charPos[k] {
			// noopではあるがアクション自体は実行可能
			return nil
		}
		if h.SwapCD > 0 {
			h.Events.Emit(event.OnActionFailed, h.active, t, param, action.SwapCD)
			return ErrActionNotReady
		}
	}

	return nil
}

// Exec はアクション t の準備状態に関わらず強制的に実行する。
// Exec の呼び出し元は、関連する場合は事前に ReadyCheck を実行済みであることが前提。
//
// この分離により、スワップCDをバイパスするスワップなど、
// 特定のアクションの強制実行が可能になる。
func (h *Handler) Exec(t action.Action, k keys.Char, param map[string]int) error {
	char := h.chars[h.active]

	// 空中専用ハンドラ: 空中の場合、次のアクションは必ず落下攻撃でなければエラー
	if h.airborne != Grounded && t != action.ActionLowPlunge && t != action.ActionHighPlunge {
		return ErrInvalidAirborneAction
	}

	var err error
	switch t {
	case action.ActionCharge: // スタミナの特殊計算が必要
		req := h.AbilStamCost(char.Index, t, param)
		h.UseStam(req, t)
		err = h.useAbility(t, param, char.ChargeAttack) //TODO: キャラクターが重撃関数内でスタミナを消費しているか確認
	case action.ActionDash:
		err = h.useAbility(t, param, char.Dash) //TODO: キャラクターがダッシュ内でスタミナを消費しているか確認
	case action.ActionJump:
		err = h.useAbility(t, param, char.Jump)
	case action.ActionWalk:
		err = h.useAbility(t, param, char.Walk)
	case action.ActionAim:
		err = h.useAbility(t, param, char.Aimed)
	case action.ActionSkill:
		err = h.useAbility(t, param, char.Skill)
	case action.ActionBurst:
		err = h.useAbility(t, param, char.Burst)
	case action.ActionAttack:
		err = h.useAbility(t, param, char.Attack)
	case action.ActionHighPlunge:
		err = h.useAbility(t, param, char.HighPlungeAttack)
		h.airborne = Grounded
	case action.ActionLowPlunge:
		err = h.useAbility(t, param, char.LowPlungeAttack)
		h.airborne = Grounded
	case action.ActionSwap:
		if h.active == h.charPos[k] {
			return ErrActionNoOp
		}
		if h.SwapCD > 0 {
			// 強制スワップを許可しているので問題ないが、念のため追加ログを出力する
			h.Log.NewEventBuildMsg(glog.LogActionEvent, h.active, "swapping ", h.chars[h.active].Base.Key.String(), " to ", h.chars[h.charPos[k]].Base.Key.String(), " (bypassed cd)").
				Write("swap_cd", h.SwapCD)
			h.SwapCD = 0
		} else {
			h.Log.NewEventBuildMsg(glog.LogActionEvent, h.active, "swapping ", h.chars[h.active].Base.Key.String(), " to ", h.chars[h.charPos[k]].Base.Key.String())
		}

		x := action.Info{
			Frames: func(action.Action) int {
				return h.Delays.Swap
			},
			AnimationLength: h.Delays.Swap,
			CanQueueAfter:   h.Delays.Swap,
			State:           action.SwapState,
		}
		x.QueueAction(h.swap(k), h.Delays.Swap)
		h.SetActionUsed(h.active, t, &x)
		h.LastAction.Type = t
		h.LastAction.Param = param
		h.LastAction.Char = h.active
	default:
		return fmt.Errorf("invalid action: %v", t)
	}
	if err != nil {
		return err
	}

	if t != action.ActionAttack {
		h.ResetAllNormalCounter()
	}

	h.Events.Emit(event.OnActionExec, h.active, t, param)

	return nil
}

var actionToEvent = map[action.Action]event.Event{
	action.ActionDash:       event.OnDash,
	action.ActionSkill:      event.OnSkill,
	action.ActionBurst:      event.OnBurst,
	action.ActionAttack:     event.OnAttack,
	action.ActionCharge:     event.OnChargeAttack,
	action.ActionLowPlunge:  event.OnPlunge,
	action.ActionHighPlunge: event.OnPlunge,
	action.ActionAim:        event.OnAimShoot,
}

func (h *Handler) useAbility(
	t action.Action,
	param map[string]int,
	f func(p map[string]int) (action.Info, error),
) error {
	state, ok := actionToEvent[t]
	if ok {
		h.Events.Emit(state)
	}
	info, err := f(param)
	if err != nil {
		return err
	}
	h.SetActionUsed(h.active, t, &info)
	if info.FramePausedOnHitlag == nil {
		info.FramePausedOnHitlag = h.ActiveChar().FramePausedOnHitlag
	}

	h.LastAction.Type = t
	h.LastAction.Param = param
	h.LastAction.Char = h.active

	h.Log.NewEventBuildMsg(
		glog.LogActionEvent,
		h.active,
		"executed ", t.String(),
	).
		Write("action", t.String()).
		Write("stam_post", h.Stam).
		Write("swap_cd_post", h.SwapCD)
	return nil
}
