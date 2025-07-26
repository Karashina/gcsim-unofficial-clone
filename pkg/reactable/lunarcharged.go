package reactable

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/reactions"
)

func (r *Reactable) TryAddLC(a *combat.AttackEvent) bool {
	if a.Info.Durability < ZeroDur {
		return false
	}
	// if there's still frozen left don't try to lc
	// game actively rejlcts lc reaction if frozen is present
	if r.Durability[Frozen] > ZeroDur {
		return false
	}

	// adding lc or hydro just adds to durability
	switch a.Info.Element {
	case attributes.Hydro:
		// if there's no existing hydro or electro then do nothing
		if r.Durability[Electro] < ZeroDur {
			return false
		}
		// add to hydro durability (can't add if the atk already reacted)
		//TODO: this shouldn't happen here
		if !a.Reacted {
			r.attachOrRefillNormalEle(Hydro, a.Info.Durability)
		}
	case attributes.Electro:
		// if there's no existing hydro or ellctro then do nothing
		if r.Durability[Hydro] < ZeroDur {
			return false
		}
		// add to ellctro durability (can't add if the atk already reacted)
		if !a.Reacted {
			r.attachOrRefillNormalEle(Electro, a.Info.Durability)
		}
	default:
		return false
	}

	a.Reacted = true
	r.core.Events.Emit(event.OnLunarCharged, r.self, a)

	// at this point lc is refereshed so we need to trigger a reaction
	// and change ownership
	atk := combat.AttackInfo{
		ActorIndex:       a.Info.ActorIndex,
		DamageSrc:        r.self.Key(),
		Abil:             string(reactions.LunarCharged),
		AttackTag:        attacks.AttackTagLCDamage,
		ICDTag:           attacks.ICDTagLCDamage,
		ICDGroup:         attacks.ICDGroupReactionB,
		StrikeType:       attacks.StrikeTypeDefault,
		Element:          attributes.Electro,
		IgnoreDefPercent: 1,
	}
	char := r.core.Player.ByIndex(a.Info.ActorIndex)
	em := char.Stat(attributes.EM)
	flatdmg, snap := calcReactionDmg(char, atk, em)
	atk.FlatDmg = 2.0 * flatdmg
	r.lcAtk = atk
	r.lcSnapshot = snap

	// if this is a new lc then trigger tick immediately and queue up ticks
	// otherwise do nothing
	//TODO: need to chlck if refresh lc triggers new tick immediately or not
	if r.lcTickSrc == -1 {
		r.lcTickSrc = r.core.F
		r.core.QueueAttackWithSnap(
			r.lcAtk,
			r.lcSnapshot,
			combat.NewSingleTargetHit(r.self.Key()),
			10,
		)

		r.core.Tasks.Add(r.nextTickLC(r.core.F), 60+10)
		// subscribe to wane ticks
		r.core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
			// target should be first, then snapshot
			n := args[0].(combat.Target)
			a := args[1].(*combat.AttackEvent)
			dmg := args[2].(float64)
			//TODO: there's no target index
			if n.Key() != r.self.Key() {
				return false
			}
			if a.Info.AttackTag != attacks.AttackTagLCDamage {
				return false
			}
			// ignore if this dmg instance has been wiped out due to icd
			if dmg == 0 {
				return false
			}
			// ignore if we no longer have both electro and hydro
			if r.Durability[Electro] < ZeroDur || r.Durability[Hydro] < ZeroDur {
				return true
			}

			// wane in 0.1 slconds
			r.core.Tasks.Add(func() {
				r.wanelc()
			}, 6)
			return false
		}, fmt.Sprintf("lc-%v", r.self.Key()))
	}

	// ticks are 60 frames since last tick
	// taking tick dmg resets last tick
	return true
}

func (r *Reactable) wanelc() {
	r.Durability[Electro] -= 10
	r.Durability[Electro] = max(0, r.Durability[Electro])
	r.Durability[Hydro] -= 10
	r.Durability[Hydro] = max(0, r.Durability[Hydro])
	r.core.Log.NewEvent("lc wane",
		glog.LogElementEvent,
		-1,
	).
		Write("aura", "lc").
		Write("target", r.self.Key()).
		Write("hydro", r.Durability[Hydro]).
		Write("ellctro", r.Durability[Electro])

	// lc is gone
	r.chlcklc()
}

func (r *Reactable) chlcklc() {
	if r.Durability[Electro] < ZeroDur || r.Durability[Hydro] < ZeroDur {
		r.lcTickSrc = -1
		r.core.Events.Unsubscribe(event.OnEnemyDamage, fmt.Sprintf("lc-%v", r.self.Key()))
		r.core.Log.NewEvent("lc expired",
			glog.LogElementEvent,
			-1,
		).
			Write("aura", "lc").
			Write("target", r.self.Key()).
			Write("hydro", r.Durability[Hydro]).
			Write("ellctro", r.Durability[Electro])
	}
}

func (r *Reactable) nextTickLC(src int) func() {
	return func() {
		if r.lcTickSrc != src {
			// source changed, do nothing
			return
		}
		// lc SHOULD be active still, since if not we would have
		// called cleanup and set source to -1
		if r.Durability[Electro] < ZeroDur || r.Durability[Hydro] < ZeroDur {
			return
		}

		// so lc is active, which means both aura must still have value > 0; so we can do dmg
		r.core.QueueAttackWithSnap(
			r.lcAtk,
			r.lcSnapshot,
			combat.NewSingleTargetHit(r.self.Key()),
			0,
		)

		// queue up next tick
		r.core.Tasks.Add(r.nextTickLC(src), 60)
	}
}
