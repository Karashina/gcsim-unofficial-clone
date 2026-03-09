package reactable

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/reactions"
)

func (r *Reactable) TryFreeze(a *combat.AttackEvent) bool {
	if a.Info.Durability < ZeroDur {
		return false
	}
	// 既に凍結している場合、2つのケースがある:
	// 1. ソースが存在するが他に共存するものがない -> 付着
	// 2. ソースが存在しないが反対側が共存する -> 凍結元素量に追加
	var consumed reactions.Durability
	switch a.Info.Element {
	case attributes.Hydro:
		// 氷が存在すれば、凍結が既に共存していても凍結を発動する
		if r.Durability[Cryo] < ZeroDur {
			return false
		}
		consumed = r.triggerFreeze(r.Durability[Cryo], a.Info.Durability)
		r.Durability[Cryo] -= consumed
		r.Durability[Cryo] = max(r.Durability[Cryo], 0)
	case attributes.Cryo:
		if r.Durability[Hydro] < ZeroDur {
			return false
		}
		consumed := r.triggerFreeze(r.Durability[Hydro], a.Info.Durability)
		r.Durability[Hydro] -= consumed
		r.Durability[Hydro] = max(r.Durability[Hydro], 0)
	default:
		// ここにあるべき
		return false
	}
	a.Reacted = true
	a.Info.Durability -= consumed
	a.Info.Durability = max(a.Info.Durability, 0)
	r.core.Events.Emit(event.OnFrozen, r.self, a)
	return true
}

func (r *Reactable) PoiseDMGCheck(a *combat.AttackEvent) bool {
	if r.Durability[Frozen] < ZeroDur {
		return false
	}
	if a.Info.StrikeType != attacks.StrikeTypeBlunt {
		return false
	}
	// 耐勢ダメージに応じて凍結元素量を削除
	r.Durability[Frozen] -= reactions.Durability(0.15 * a.Info.PoiseDMG)
	r.checkFreeze()
	return true
}

func (r *Reactable) ShatterCheck(a *combat.AttackEvent) bool {
	if r.Durability[Frozen] < ZeroDur {
		return false
	}
	if a.Info.StrikeType != attacks.StrikeTypeBlunt && a.Info.Element != attributes.Geo {
		return false
	}
	// 可能なら凍結ゲージ200を削除
	r.Durability[Frozen] -= 200
	r.checkFreeze()

	r.core.Events.Emit(event.OnShatter, r.self, a)

	// 砕氷攻撃の0.2秒GCD
	if !(r.shatterGCD != -1 && r.core.F < r.shatterGCD) {
		r.shatterGCD = r.core.F + 0.2*60
		// 砕氷攻撃を発動
		ai := combat.AttackInfo{
			ActorIndex:       a.Info.ActorIndex,
			DamageSrc:        r.self.Key(),
			Abil:             string(reactions.Shatter),
			AttackTag:        attacks.AttackTagShatter,
			ICDTag:           attacks.ICDTagShatter,
			ICDGroup:         attacks.ICDGroupReactionA,
			StrikeType:       attacks.StrikeTypeDefault,
			Element:          attributes.Physical,
			IgnoreDefPercent: 1,
		}
		char := r.core.Player.ByIndex(a.Info.ActorIndex)
		em := char.Stat(attributes.EM)
		flatdmg, snap := calcReactionDmg(char, ai, em)
		ai.FlatDmg = 3.0 * flatdmg
		// 砕氷は自身への攻撃
		r.core.QueueAttackWithSnap(
			ai,
			snap,
			combat.NewSingleTargetHit(r.self.Key()),
			0,
		)
	}

	return true
}

// triggerFreeze は凍結元素量に追加し、消費された元素量を返す
func (r *Reactable) triggerFreeze(a, b reactions.Durability) reactions.Durability {
	d := min(a, b)
	if r.FreezeResist >= 1 {
		return d
	}
	// triggerFreeze は addDurability のみ行い、減衰速度には触れない
	r.attachOverlap(Frozen, 2*d, ZeroDur)
	return d
}

func (r *Reactable) checkFreeze() {
	if r.Durability[Frozen] <= ZeroDur {
		r.Durability[Frozen] = 0
		r.core.Events.Emit(event.OnAuraDurabilityDepleted, r.self, attributes.Frozen)
		// バブルを割るためだけにここで別の攻撃を発動する >.>
		ai := combat.AttackInfo{
			ActorIndex:  0,
			DamageSrc:   r.self.Key(),
			Abil:        "Freeze Broken",
			AttackTag:   attacks.AttackTagNone,
			ICDTag:      attacks.ICDTagNone,
			ICDGroup:    attacks.ICDGroupDefault,
			StrikeType:  attacks.StrikeTypeDefault,
			Element:     attributes.NoElement,
			SourceIsSim: true,
			DoNotLog:    true,
		}
		//TODO: 攻撃を1フレーム遅延させてOKか？
		r.core.QueueAttack(ai, combat.NewSingleTargetHit(r.self.Key()), -1, 0)
	}
}
