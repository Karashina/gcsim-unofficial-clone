package reactable

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/reactions"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

func calcSwirlAtkDurability(consumed, src reactions.Durability) reactions.Durability {
	if consumed < src {
		return 1.25*(0.5*consumed-1) + 25
	}
	return 1.25*(src-1) + 25
}

func (r *Reactable) queueSwirl(rt reactions.ReactionType, ele attributes.Element, tag attacks.AttackTag, icd attacks.ICDTag, dur reactions.Durability, charIndex int) {
	// 拡散は2つの攻撃を発動；ゲージなしの自身への攻撃と
	// ゲージありのAoE攻撃
	ai := combat.AttackInfo{
		ActorIndex:       charIndex,
		DamageSrc:        r.self.Key(),
		Abil:             string(rt),
		AttackTag:        tag,
		ICDTag:           icd,
		ICDGroup:         attacks.ICDGroupReactionA,
		StrikeType:       attacks.StrikeTypeDefault,
		Element:          ele,
		IgnoreDefPercent: 1,
	}
	char := r.core.Player.ByIndex(charIndex)
	em := char.Stat(attributes.EM)
	flatdmg, snap := calcReactionDmg(char, ai, em)
	ai.FlatDmg = 0.6 * flatdmg
	// 最初の攻撃は自身へ、ヒットボックスなし
	r.core.QueueAttackWithSnap(
		ai,
		snap,
		combat.NewSingleTargetHit(r.self.Key()),
		1,
	)
	// 次はAoE - 水拡散はAoEダメージを与えず、元素を広げるだけ
	if ele == attributes.Hydro {
		ai.FlatDmg = 0
	}
	ai.Durability = dur
	ai.Abil = string(rt) + " (aoe)"
	ap := combat.NewCircleHitOnTarget(r.self, nil, 5)
	ap.IgnoredKeys = []targets.TargetKey{r.self.Key()}
	r.core.QueueAttackWithSnap(
		ai,
		snap,
		ap,
		1+4, // 4f after self swirl
	)
}

func (r *Reactable) TrySwirlElectro(a *combat.AttackEvent) bool {
	if a.Info.Durability < ZeroDur {
		return false
	}
	if r.Durability[Electro] < ZeroDur {
		return false
	}
	rd := r.reduce(attributes.Electro, a.Info.Durability, 0.5)
	atkDur := calcSwirlAtkDurability(rd, a.Info.Durability)
	a.Info.Durability -= rd
	a.Reacted = true
	// まず攻撃をキューに追加
	r.core.Events.Emit(event.OnSwirlElectro, r.self, a)

	// 雷拡散攻撃の0.1秒GCD
	if !(r.swirlElectroGCD != -1 && r.core.F < r.swirlElectroGCD) {
		r.swirlElectroGCD = r.core.F + 0.1*60
		r.queueSwirl(
			reactions.SwirlElectro,
			attributes.Electro,
			attacks.AttackTagSwirlElectro,
			attacks.ICDTagSwirlElectro,
			atkDur,
			a.Info.ActorIndex,
		)
	}

	// この時点で元素量が残っている場合、感電反応のために
	// 水の存在をチェックする必要がある
	if a.Info.Durability > ZeroDur && r.Durability[Hydro] > ZeroDur {
		// 水拡散を発動
		r.TrySwirlHydro(a)
		// 感電反応のクリーンアップをチェック
		r.checkEC()
	}
	return true
}

func (r *Reactable) TrySwirlHydro(a *combat.AttackEvent) bool {
	if a.Info.Durability < ZeroDur {
		return false
	}
	if r.Durability[Hydro] < ZeroDur {
		return false
	}
	rd := r.reduce(attributes.Hydro, a.Info.Durability, 0.5)
	atkDur := calcSwirlAtkDurability(rd, a.Info.Durability)
	a.Info.Durability -= rd
	a.Reacted = true
	// まず攻撃をキューに追加
	r.core.Events.Emit(event.OnSwirlHydro, r.self, a)

	// 水拡散攻撃の0.1秒GCD
	if !(r.swirlHydroGCD != -1 && r.core.F < r.swirlHydroGCD) {
		r.swirlHydroGCD = r.core.F + 0.1*60
		r.queueSwirl(
			reactions.SwirlHydro,
			attributes.Hydro,
			attacks.AttackTagSwirlHydro,
			attacks.ICDTagSwirlHydro,
			atkDur,
			a.Info.ActorIndex,
		)
	}

	return true
}

func (r *Reactable) TrySwirlCryo(a *combat.AttackEvent) bool {
	if a.Info.Durability < ZeroDur {
		return false
	}
	if r.Durability[Cryo] < ZeroDur {
		return false
	}
	rd := r.reduce(attributes.Cryo, a.Info.Durability, 0.5)
	atkDur := calcSwirlAtkDurability(rd, a.Info.Durability)
	a.Info.Durability -= rd
	a.Reacted = true
	// まず攻撃をキューに追加
	r.core.Events.Emit(event.OnSwirlCryo, r.self, a)

	// 0.1s gcd on swirl cryo attack
	if !(r.swirlCryoGCD != -1 && r.core.F < r.swirlCryoGCD) {
		r.swirlCryoGCD = r.core.F + 0.1*60
		r.queueSwirl(
			reactions.SwirlCryo,
			attributes.Cryo,
			attacks.AttackTagSwirlCryo,
			attacks.ICDTagSwirlCryo,
			atkDur,
			a.Info.ActorIndex,
		)
	}

	return true
}

func (r *Reactable) TrySwirlPyro(a *combat.AttackEvent) bool {
	if a.Info.Durability < ZeroDur {
		return false
	}
	if r.Durability[Pyro] < ZeroDur {
		return false
	}
	rd := r.reduce(attributes.Pyro, a.Info.Durability, 0.5)
	atkDur := calcSwirlAtkDurability(rd, a.Info.Durability)
	a.Info.Durability -= rd
	a.Reacted = true
	r.burningCheck()
	// まず攻撃をキューに追加
	r.core.Events.Emit(event.OnSwirlPyro, r.self, a)

	// 炎拡散攻撃の0.1秒GCD
	if !(r.swirlPyroGCD != -1 && r.core.F < r.swirlPyroGCD) {
		r.swirlPyroGCD = r.core.F + 0.1*60
		r.queueSwirl(
			reactions.SwirlPyro,
			attributes.Pyro,
			attacks.AttackTagSwirlPyro,
			attacks.ICDTagSwirlPyro,
			atkDur,
			a.Info.ActorIndex,
		)
	}

	return true
}

func (r *Reactable) TrySwirlFrozen(a *combat.AttackEvent) bool {
	if a.Info.Durability < ZeroDur {
		return false
	}
	if r.Durability[Frozen] < ZeroDur {
		return false
	}
	rd := r.reduce(attributes.Frozen, a.Info.Durability, 0.5)
	atkDur := calcSwirlAtkDurability(rd, a.Info.Durability)
	a.Info.Durability -= rd
	a.Reacted = true
	// まず攻撃をキューに追加
	r.core.Events.Emit(event.OnSwirlCryo, r.self, a)

	// 0.1s gcd on swirl cryo attack
	if !(r.swirlCryoGCD != -1 && r.core.F < r.swirlCryoGCD) {
		r.swirlCryoGCD = r.core.F + 0.1*60
		r.queueSwirl(
			reactions.SwirlCryo,
			attributes.Cryo,
			attacks.AttackTagSwirlCryo,
			attacks.ICDTagSwirlCryo,
			atkDur,
			a.Info.ActorIndex,
		)
	}

	r.checkFreeze()
	return true
}
