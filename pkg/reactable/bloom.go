package reactable

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/reactions"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/gadget"
)

const DendroCoreDelay = 30

func (r *Reactable) TryBloom(a *combat.AttackEvent) bool {
	// can be hydro bloom, dendro bloom, or quicken bloom
	if a.Info.Durability < ZeroDur {
		return false
	}
	var consumed reactions.Durability
	switch a.Info.Element {
	case attributes.Hydro:
		// this part is annoying. bloom will happen if any of the dendro like aura is present
		// so we gotta check for all 3...
		switch {
		case r.Durability[Dendro] > ZeroDur:
		case r.Durability[Quicken] > ZeroDur:
		case r.Durability[BurningFuel] > ZeroDur:
		default:
			return false
		}
		// reduce only check for one element so have to call twice to check for quicken as well
		consumed = r.reduce(attributes.Dendro, a.Info.Durability, 0.5)
		f := r.reduce(attributes.Quicken, a.Info.Durability, 0.5)
		if f > consumed {
			consumed = f
		}
	case attributes.Dendro:
		if r.Durability[Hydro] < ZeroDur {
			return false
		}
		consumed = r.reduce(attributes.Hydro, a.Info.Durability, 2)
	default:
		return false
	}
	a.Info.Durability -= consumed
	a.Info.Durability = max(a.Info.Durability, 0)
	a.Reacted = true

	r.addBloomGadget(a)
	if r.core.Player.ByIndex(a.Info.ActorIndex).StatusIsActive("LB-Key") {
		r.core.Events.Emit(event.OnLunarBloom, r.self, a)
	}
	r.core.Events.Emit(event.OnBloom, r.self, a)
	return true
}

// this function should only be called after a catalyze reaction (queued to the end of current frame)
// this reaction will check if any hydro exists and if so trigger a bloom reaction
func (r *Reactable) tryQuickenBloom(a *combat.AttackEvent) {
	if r.Durability[Quicken] < ZeroDur {
		// this should be a sanity check; should not happen realistically unless something wipes off
		// the quicken immediately (same frame) after catalyze
		return
	}
	if r.Durability[Hydro] < ZeroDur {
		return
	}
	avail := r.Durability[Quicken]
	consumed := r.reduce(attributes.Hydro, avail, 2)
	r.Durability[Quicken] -= consumed

	r.addBloomGadget(a)
	r.core.Events.Emit(event.OnBloom, r.self, a)
}

type DendroCore struct {
	*gadget.Gadget
	srcFrame  int
	CharIndex int
	// If true, this dendro core is a Seed of Deceit (Nefer)
	IsSeed bool
}

func (r *Reactable) addBloomGadget(a *combat.AttackEvent) {
	// determine if this should be a Seed of Deceit at scheduling time to avoid timing races
	isSeed := false
	for _, ch := range r.core.Player.Chars() {
		if ch.StatusIsActive("nefer-seed-convert") {
			isSeed = true
			break
		}
	}
	// capture isSeed in closure
	capturedIsSeed := isSeed
	r.core.Tasks.Add(func() {
		t := NewDendroCore(r.core, r.self.Shape(), a, capturedIsSeed)
		r.core.Combat.AddGadget(t)
		r.core.Events.Emit(event.OnDendroCore, t, a)
		r.core.Log.NewEvent(
			"dendro core spawned",
			glog.LogElementEvent,
			a.Info.ActorIndex,
		).
			Write("src", t.Src()).
			Write("expiry", r.core.F+t.Duration).
			Write("is_seed", capturedIsSeed)
	}, DendroCoreDelay)
}

func NewDendroCore(c *core.Core, shp geometry.Shape, a *combat.AttackEvent, isSeed bool) *DendroCore {
	s := &DendroCore{
		srcFrame:  c.F,
		CharIndex: a.Info.ActorIndex,
		IsSeed:    isSeed,
	}

	circ, ok := shp.(*geometry.Circle)
	if !ok {
		panic("rectangle target hurtbox is not supported for dendro core spawning")
	}

	// for simplicity, seeds spawn randomly at radius + 0.5
	r := circ.Radius() + 0.5
	s.Gadget = gadget.New(c, geometry.CalcRandomPointFromCenter(circ.Pos(), r, r, c.Rand), 2, combat.GadgetTypDendroCore)
	s.Gadget.Duration = 300 // ??

	// Log creation details for debugging Seed of Deceit behavior
	c.Log.NewEvent(
		"new dendro core created",
		glog.LogElementEvent,
		a.Info.ActorIndex,
	).Write("is_seed", isSeed).
		Write("src_frame", s.srcFrame)

	char := s.Core.Player.ByIndex(a.Info.ActorIndex)

	explode := func(reason string) func() {
		return func() {
			s.Core.Tasks.Add(func() {
				// Log an explode attempt and current IsSeed state
				s.Core.Log.NewEvent(
					"dendro core explode attempt",
					glog.LogElementEvent,
					char.Index,
				).Write("is_seed", s.IsSeed).
					Write("src", s.Src())

				if s.IsSeed {
					// Seeds of Deceit do not explode
					return
				}
				// bloom attack
				ai, snap := NewBloomAttack(char, s, nil)
				ap := combat.NewCircleHitOnTarget(s, nil, 5)
				c.QueueAttackWithSnap(ai, snap, ap, 0)

				// self damage
				ai.Abil += reactions.SelfDamageSuffix
				ai.FlatDmg = 0.05 * ai.FlatDmg
				ap.SkipTargets[targets.TargettablePlayer] = false
				ap.SkipTargets[targets.TargettableEnemy] = true
				ap.SkipTargets[targets.TargettableGadget] = true
				c.QueueAttackWithSnap(ai, snap, ap, 0)

				c.Log.NewEvent(
					"dendro core "+reason,
					glog.LogElementEvent,
					char.Index,
				).Write("src", s.Src())
			}, 1)
		}
	}
	if !isSeed {
		s.Gadget.OnExpiry = explode("expired")
		s.Gadget.OnKill = explode("killed")
	} else {
		// Seeds of Deceit do not burst or cause reactions on expiry/kill
		s.Gadget.OnExpiry = nil
		s.Gadget.OnKill = nil
	}

	return s
}

func (s *DendroCore) Tick() {
	// this is needed since gadget tick
	s.Gadget.Tick()
}

func (s *DendroCore) HandleAttack(atk *combat.AttackEvent) float64 {
	s.Core.Events.Emit(event.OnGadgetHit, s, atk)
	s.Attack(atk, nil)
	return 0
}

func (s *DendroCore) Attack(atk *combat.AttackEvent, evt glog.Event) (float64, bool) {
	if atk.Info.Durability < ZeroDur {
		return 0, false
	}

	char := s.Core.Player.ByIndex(atk.Info.ActorIndex)
	// only contact with pyro/electro to trigger burgeon/hyperbloom accordingly
	switch atk.Info.Element {
	case attributes.Electro:
		if s.IsSeed {
			// Seeds of Deceit cannot trigger Hyperbloom
			return 0, false
		}
		// trigger hyperbloom targets the nearest enemy
		// it can also do damage to player in small aoe
		s.Core.Tasks.Add(func() {
			ai, snap := NewHyperbloomAttack(char, s)
			// queue dmg nearest enemy within radius 15
			enemy := s.Core.Combat.ClosestEnemyWithinArea(combat.NewCircleHitOnTarget(s.Gadget, nil, 15), nil)
			if enemy != nil {
				ap := combat.NewCircleHitOnTarget(enemy, nil, 1)
				s.Core.QueueAttackWithSnap(ai, snap, ap, 0)

				// also queue self damage
				ai.Abil += reactions.SelfDamageSuffix
				ai.FlatDmg = 0.05 * ai.FlatDmg
				ap.SkipTargets[targets.TargettablePlayer] = false
				ap.SkipTargets[targets.TargettableEnemy] = true
				ap.SkipTargets[targets.TargettableGadget] = true
				s.Core.QueueAttackWithSnap(ai, snap, ap, 0)
			}
		}, 60)

		s.Gadget.OnKill = nil
		s.Gadget.Kill()
		s.Core.Events.Emit(event.OnHyperbloom, s, atk)
		s.Core.Log.NewEvent(
			"hyperbloom triggered",
			glog.LogElementEvent,
			char.Index,
		).
			Write("dendro_core_char", s.CharIndex).
			Write("dendro_core_src", s.Gadget.Src())
	case attributes.Pyro:
		if s.IsSeed {
			// Seeds of Deceit cannot trigger Burgeon
			return 0, false
		}
		// trigger burgeon, aoe dendro damage
		// self damage
		s.Core.Tasks.Add(func() {
			ai, snap := NewBurgeonAttack(char, s)
			ap := combat.NewCircleHitOnTarget(s, nil, 5)
			s.Core.QueueAttackWithSnap(ai, snap, ap, 0)

			// queue self damage
			ai.Abil += reactions.SelfDamageSuffix
			ai.FlatDmg = 0.05 * ai.FlatDmg
			ap.SkipTargets[targets.TargettablePlayer] = false
			ap.SkipTargets[targets.TargettableEnemy] = true
			ap.SkipTargets[targets.TargettableGadget] = true
			s.Core.QueueAttackWithSnap(ai, snap, ap, 0)
		}, 1)

		s.Gadget.OnKill = nil
		s.Gadget.Kill()
		s.Core.Events.Emit(event.OnBurgeon, s, atk)
		s.Core.Log.NewEvent(
			"burgeon triggered",
			glog.LogElementEvent,
			char.Index,
		).
			Write("dendro_core_char", s.CharIndex).
			Write("dendro_core_src", s.Gadget.Src())
	default:
		return 0, false
	}

	return 0, false
}

const (
	BloomMultiplier      = 2
	BurgeonMultiplier    = 3
	HyperbloomMultiplier = 3
)

func NewBloomAttack(char *character.CharWrapper, src combat.Target, modify func(*combat.AttackInfo)) (combat.AttackInfo, combat.Snapshot) {
	em := char.Stat(attributes.EM)
	ai := combat.AttackInfo{
		ActorIndex:       char.Index,
		DamageSrc:        src.Key(),
		Element:          attributes.Dendro,
		AttackTag:        attacks.AttackTagBloom,
		ICDTag:           attacks.ICDTagBloomDamage,
		ICDGroup:         attacks.ICDGroupReactionA,
		StrikeType:       attacks.StrikeTypeDefault,
		Abil:             string(reactions.Bloom),
		IgnoreDefPercent: 1,
	}
	if modify != nil {
		modify(&ai)
	}
	flatdmg, snap := calcReactionDmg(char, ai, em)
	ai.FlatDmg = BloomMultiplier * flatdmg
	return ai, snap
}

func NewBurgeonAttack(char *character.CharWrapper, src combat.Target) (combat.AttackInfo, combat.Snapshot) {
	em := char.Stat(attributes.EM)
	ai := combat.AttackInfo{
		ActorIndex:       char.Index,
		DamageSrc:        src.Key(),
		Element:          attributes.Dendro,
		AttackTag:        attacks.AttackTagBurgeon,
		ICDTag:           attacks.ICDTagBurgeonDamage,
		ICDGroup:         attacks.ICDGroupReactionA,
		StrikeType:       attacks.StrikeTypeDefault,
		Abil:             string(reactions.Burgeon),
		IgnoreDefPercent: 1,
	}
	flatdmg, snap := calcReactionDmg(char, ai, em)
	ai.FlatDmg = BurgeonMultiplier * flatdmg
	return ai, snap
}

func NewHyperbloomAttack(char *character.CharWrapper, src combat.Target) (combat.AttackInfo, combat.Snapshot) {
	em := char.Stat(attributes.EM)
	ai := combat.AttackInfo{
		ActorIndex:       char.Index,
		DamageSrc:        src.Key(),
		Element:          attributes.Dendro,
		AttackTag:        attacks.AttackTagHyperbloom,
		ICDTag:           attacks.ICDTagHyperbloomDamage,
		ICDGroup:         attacks.ICDGroupReactionA,
		StrikeType:       attacks.StrikeTypeDefault,
		Abil:             string(reactions.Hyperbloom),
		IgnoreDefPercent: 1,
	}
	flatdmg, snap := calcReactionDmg(char, ai, em)
	ai.FlatDmg = HyperbloomMultiplier * flatdmg
	return ai, snap
}

func (s *DendroCore) SetDirection(trg geometry.Point) {}
func (s *DendroCore) SetDirectionToClosestEnemy()     {}
func (s *DendroCore) CalcTempDirection(trg geometry.Point) geometry.Point {
	return geometry.DefaultDirection()
}
