package reactable

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/reactions"
)

func (r *Reactable) TryVaporize(a *combat.AttackEvent) bool {
	if a.Info.Durability < ZeroDur {
		return false
	}
	var consumed reactions.Durability
	switch a.Info.Element {
	case attributes.Pyro:
		// 水が存在することを確認
		if r.Durability[Hydro] < ZeroDur {
			return false
		}
		// 凍結がまだ残っている場合、蒸発を試みない
		// ゲームは凍結が存在する場合蒸発反応を積極的に拒否する
		if r.Durability[Frozen] > ZeroDur {
			return false
		}
		consumed = r.reduce(attributes.Hydro, a.Info.Durability, .5)
		a.Info.AmpMult = 1.5
	case attributes.Hydro:
		// 炎が存在することを確認；炎との共存はまだない
		if r.Durability[Pyro] < ZeroDur && r.Durability[Burning] < ZeroDur {
			return false
		}
		consumed = r.reduce(attributes.Pyro, a.Info.Durability, 2)
		a.Info.AmpMult = 2
		r.burningCheck()
	default:
		// ここにあるべき
		return false
	}
	// 他に反応するものはないはず
	a.Info.Durability -= consumed
	a.Info.Durability = max(a.Info.Durability, 0)
	a.Reacted = true
	a.Info.Amped = true
	a.Info.AmpType = reactions.Vaporize
	r.core.Events.Emit(event.OnVaporize, r.self, a)
	return true
}
