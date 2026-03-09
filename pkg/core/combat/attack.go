package combat

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

// attack は攻撃が命中した場合にtrueを返す
func (h *Handler) attack(t Target, a *AttackEvent) (float64, bool) {
	willHit, reason := t.AttackWillLand(a.Pattern)
	if !willHit {
		// Guobaなどのメイン表示を散らかさないよう、ターゲットログを"Sim"イベントログに移動
		// 「Fischl A4は単体攻撃なのでターゲット2-4には命中しない」のような自明なことも含む
		// TODO: このための別のログイベントセットを追加すべきか？
		if h.Debug && t.Type() != targets.TargettablePlayer {
			h.Log.NewEventBuildMsg(glog.LogDebugEvent, a.Info.ActorIndex, "skipped ", a.Info.Abil, " ", reason).
				Write("attack_tag", a.Info.AttackTag).
				Write("applied_ele", a.Info.Element).
				Write("dur", a.Info.Durability).
				Write("target", t.Key()).
				Write("geometry.Shape", a.Pattern.Shape.String())
		}
		return 0, false
	}
	// まずコピーを作成
	cpy := *a
	dmg := t.HandleAttack(&cpy)
	return dmg, true
}

func (h *Handler) ApplyAttack(a *AttackEvent) float64 {
	h.Events.Emit(event.OnApplyAttack, a)
	// died := false
	var total float64
	var landed bool
	// プレイヤーをチェック
	if !a.Pattern.SkipTargets[targets.TargettablePlayer] {
		//TODO: プレイヤーへの攻撃はヒットラグを生成しないはずなので、ここではlandedをチェックしない
		h.attack(h.player, a)
	}
	// 敵をチェック
	if !a.Pattern.SkipTargets[targets.TargettableEnemy] {
		for _, v := range h.enemies {
			if v == nil {
				continue
			}
			if !v.IsAlive() {
				continue
			}
			a, l := h.attack(v, a)
			total += a
			if l {
				landed = true
			}
		}
	}
	// ガジェットをチェック
	if !a.Pattern.SkipTargets[targets.TargettableGadget] {
		for i := 0; i < len(h.gadgets); i++ {
			// 死亡してクリーンアップがまだのガジェットがありえるので、ここで安全性チェック
			if h.gadgets[i] == nil {
				continue
			}
			h.attack(h.gadgets[i], a)
		}
	}
	// アクターにヒットラグを追加（設置物の場合は無視）
	if h.EnableHitlag && landed && !a.Info.IsDeployable {
		dur := a.Info.HitlagHaltFrames
		if h.DefHalt && a.Info.CanBeDefenseHalted {
			dur += 3.6 // 0.06
		}
		if dur > 0 {
			h.Team.ApplyHitlag(a.Info.ActorIndex, a.Info.HitlagFactor, dur)
			if h.Debug {
				h.Log.NewEvent(fmt.Sprintf("%v applying hitlag: %.3f", a.Info.Abil, dur), glog.LogHitlagEvent, a.Info.ActorIndex).
					Write("duration", dur).
					Write("factor", a.Info.HitlagFactor)
			}
		}
	}
	return total
}
