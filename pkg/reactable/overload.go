package reactable

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/reactions"
)

func (r *Reactable) TryOverload(a *combat.AttackEvent) bool {
	if a.Info.Durability < ZeroDur {
		return false
	}
	var consumed reactions.Durability
	switch a.Info.Element {
	case attributes.Electro:
		// 炎が必要；炎は共存できない（今のところ）ので数を無視してOK？
		if r.Durability[Pyro] < ZeroDur && r.Durability[Burning] < ZeroDur {
			return false
		}
		// 元素量を減少；なくなるか残るか；実際に反応した量は気にしない
		consumed = r.reduce(attributes.Pyro, a.Info.Durability, 1)
		r.burningCheck()
	case attributes.Pyro:
		// 雷が必要；感電反応に注意？
		if r.Durability[Electro] < ZeroDur {
			return false
		}
		consumed = r.reduce(attributes.Electro, a.Info.Durability, 1)
	default:
		// ここにあるべき
		return false
	}
	a.Info.Durability -= consumed
	a.Info.Durability = max(a.Info.Durability, 0)
	a.Reacted = true

	// 攻撃がキューされる前にイベントを発動。他のアクションが修正する時間を与える
	r.core.Events.Emit(event.OnOverload, r.self, a)

	// 過負荷攻撃の0.1秒GCD
	if !(r.overloadGCD != -1 && r.core.F < r.overloadGCD) {
		r.overloadGCD = r.core.F + 0.1*60
		// 過負荷攻撃を発動
		atk := combat.AttackInfo{
			ActorIndex:       a.Info.ActorIndex,
			DamageSrc:        r.self.Key(),
			Abil:             string(reactions.Overload),
			AttackTag:        attacks.AttackTagOverloadDamage,
			ICDTag:           attacks.ICDTagOverloadDamage,
			ICDGroup:         attacks.ICDGroupReactionB,
			StrikeType:       attacks.StrikeTypeBlunt,
			PoiseDMG:         90,
			Element:          attributes.Pyro,
			IgnoreDefPercent: 1,
		}
		char := r.core.Player.ByIndex(a.Info.ActorIndex)
		em := char.Stat(attributes.EM)
		flatdmg, snap := calcReactionDmg(char, atk, em)
		atk.FlatDmg = 2.75 * flatdmg
		r.core.QueueAttackWithSnap(atk, snap, combat.NewCircleHitOnTarget(r.self, nil, 3), 1)
	}

	return true
}
