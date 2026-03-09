package reactable

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/reactions"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

func (r *Reactable) TryBurning(a *combat.AttackEvent) bool {
	if a.Info.Durability < ZeroDur {
		return false
	}

	dendroDur := r.Durability[Dendro]

	// 炎または草を追加すると元素量が増加するだけ
	switch a.Info.Element {
	case attributes.Pyro:
		// 既存の炎/燃焼または草/激化がなければ何もしない
		if r.Durability[Dendro] < ZeroDur && r.Durability[Quicken] < ZeroDur {
			return false
		}
		// 炎元素量に追加
		// r.attachOrRefillNormalEle(ModifierPyro, a.Info.Durability)
	case attributes.Dendro:
		// 既存の炎/燃焼または草/激化がなければ何もしない
		if r.Durability[Pyro] < ZeroDur && r.Durability[Burning] < ZeroDur {
			return false
		}
		dendroDur = max(dendroDur, a.Info.Durability*0.8)
		// 草元素量に追加
		// r.attachOrRefillNormalEle(ModifierDendro, a.Info.Durability)
	default:
		return false
	}
	// a.Reacted = true

	if r.Durability[BurningFuel] < ZeroDur {
		r.attachBurningFuel(max(dendroDur, r.Durability[Quicken]), 1)
		r.attachBurning()

		r.core.Events.Emit(event.OnBurning, r.self, a)
		r.calcBurningDmg(a)

		if r.burningTickSrc == -1 {
			r.burningTickSrc = r.core.F
			if t, ok := r.self.(Enemy); ok {
				// 燃焼ティックをキュー
				t.QueueEnemyTask(r.nextBurningTick(r.core.F, 1, t), 15)
			}
		}
		return true
	}
	// 燃焼燃料を上書きし、燃焼ダメージを再計算
	if a.Info.Element == attributes.Dendro {
		r.attachBurningFuel(a.Info.Durability, 0.8)
	}
	r.calcBurningDmg(a)

	return false
}

func (r *Reactable) attachBurningFuel(dur, mult reactions.Durability) {
	// 燃焼燃料は常に上書きされる
	r.Durability[BurningFuel] = mult * dur
	decayRate := mult * dur / (6*dur + 420)
	if decayRate < 10.0/60.0 {
		decayRate = 10.0 / 60.0
	}
	r.DecayRate[BurningFuel] = decayRate
}

func (r *Reactable) calcBurningDmg(a *combat.AttackEvent) {
	atk := combat.AttackInfo{
		ActorIndex:       a.Info.ActorIndex,
		DamageSrc:        r.self.Key(),
		Abil:             string(reactions.Burning),
		AttackTag:        attacks.AttackTagBurningDamage,
		ICDTag:           attacks.ICDTagBurningDamage,
		ICDGroup:         attacks.ICDGroupBurning,
		StrikeType:       attacks.StrikeTypeDefault,
		Element:          attributes.Pyro,
		Durability:       25,
		IgnoreDefPercent: 1,
	}
	char := r.core.Player.ByIndex(a.Info.ActorIndex)
	em := char.Stat(attributes.EM)
	flatdmg, snap := calcReactionDmg(char, atk, em)
	atk.FlatDmg = 0.25 * flatdmg
	r.burningAtk = atk
	r.burningSnapshot = snap
}

func (r *Reactable) nextBurningTick(src, counter int, t Enemy) func() {
	return func() {
		if r.burningTickSrc != src {
			// ソースが変更されたので何もしない
			return
		}
		// 燃焼はまだアクティブであるはず。そうでなければクリーンアップが
		// 呼ばれてソースが-1に設定されているはず
		if r.Durability[BurningFuel] < ZeroDur || r.Durability[Burning] < ZeroDur {
			return
		}
		// 燃焼がアクティブであるため、両オーラの値が0より大きいはずなのでダメージを与える
		if counter != 9 {
			// 9番目のティックはスキップ（HoYoverseのスパゲッティコードのため）
			ai := r.burningAtk
			ap := combat.NewCircleHitOnTarget(r.self, nil, 1)
			r.core.QueueAttackWithSnap(
				ai,
				r.burningSnapshot,
				ap,
				0,
			)
			// 自傷ダメージ
			ai.Abil += reactions.SelfDamageSuffix
			ap.SkipTargets[targets.TargettablePlayer] = false
			ap.SkipTargets[targets.TargettableEnemy] = true
			ap.SkipTargets[targets.TargettableGadget] = true
			r.core.QueueAttackWithSnap(
				ai,
				r.burningSnapshot,
				ap,
				0,
			)
		}
		counter++
		// 次のティックをキュー
		t.QueueEnemyTask(r.nextBurningTick(src, counter, t), 15)
	}
}

// burningCheck は燃焼がアクティブでなくなった場合に修飾子をパージする
func (r *Reactable) burningCheck() {
	if r.Durability[Burning] < ZeroDur && r.Durability[BurningFuel] > ZeroDur {
		// 燃焼ティックはもう発生しない
		r.burningTickSrc = -1
		// 燃焼燃料を除去
		r.Durability[BurningFuel] = 0
		r.DecayRate[BurningFuel] = 0
	}
}
