package reactable

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/reactions"
)

func (r *Reactable) TrySuperconduct(a *combat.AttackEvent) bool {
	if a.Info.Durability < ZeroDur {
		return false
	}
	// 非凍結の超電導用
	if r.Durability[Frozen] >= ZeroDur {
		return false
	}
	var consumed reactions.Durability
	switch a.Info.Element {
	case attributes.Electro:
		if r.Durability[Cryo] < ZeroDur {
			return false
		}
		consumed = r.reduce(attributes.Cryo, a.Info.Durability, 1)
	case attributes.Cryo:
		// 感電反応が存在する可能性がある
		if r.Durability[Electro] < ZeroDur {
			return false
		}
		consumed = r.reduce(attributes.Electro, a.Info.Durability, 1)
	default:
		return false
	}

	a.Info.Durability -= consumed
	a.Info.Durability = max(a.Info.Durability, 0)
	a.Reacted = true
	r.queueSuperconduct(a)
	return true
}

func (r *Reactable) TryFrozenSuperconduct(a *combat.AttackEvent) bool {
	if a.Info.Durability < ZeroDur {
		return false
	}
	// 凍結がある場合の超電導
	if r.Durability[Frozen] < ZeroDur {
		return false
	}
	switch a.Info.Element {
	case attributes.Electro:
		//TODO: 前提としてまず氷を減少させ、ソース元素量が残っていれば
		// 凍結を減少させる。おれでも超電導反応は1回だけであることに注意
		a.Info.Durability -= r.reduce(attributes.Cryo, a.Info.Durability, 1)
		r.reduce(attributes.Frozen, a.Info.Durability, 1)
		a.Info.Durability = 0
		a.Reacted = true
	default:
		return false
	}

	r.queueSuperconduct(a)

	return false
}

func (r *Reactable) queueSuperconduct(a *combat.AttackEvent) {
	r.core.Events.Emit(event.OnSuperconduct, r.self, a)

	// 超電導攻撃の0.1秒GCD
	if r.superconductGCD != -1 && r.core.F < r.superconductGCD {
		return
	}
	r.superconductGCD = r.core.F + 0.1*60

	// 超電導攻撃
	atk := combat.AttackInfo{
		ActorIndex:       a.Info.ActorIndex,
		DamageSrc:        r.self.Key(),
		Abil:             string(reactions.Superconduct),
		AttackTag:        attacks.AttackTagSuperconductDamage,
		ICDTag:           attacks.ICDTagSuperconductDamage,
		ICDGroup:         attacks.ICDGroupReactionA,
		StrikeType:       attacks.StrikeTypeDefault,
		Element:          attributes.Cryo,
		IgnoreDefPercent: 1,
	}
	char := r.core.Player.ByIndex(a.Info.ActorIndex)
	em := char.Stat(attributes.EM)
	flatdmg, snap := calcReactionDmg(char, atk, em)
	atk.FlatDmg = 1.5 * flatdmg
	r.core.QueueAttackWithSnap(atk, snap, combat.NewCircleHitOnTarget(r.self, nil, 3), 1)
}
