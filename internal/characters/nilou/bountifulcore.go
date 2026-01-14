package nilou

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/reactions"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/gadget"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/reactable"
)

type BountifulCore struct {
	srcFrame int
	*gadget.Gadget
}

func newBountifulCore(c *core.Core, p geometry.Point, a *combat.AttackEvent) *BountifulCore {
	b := &BountifulCore{
		srcFrame: c.F,
	}

	b.Gadget = gadget.New(c, p, 2, combat.GadgetTypDendroCore)
	b.Gadget.Duration = 0.4 * 60

	char := b.Core.Player.ByIndex(a.Info.ActorIndex)
	explode := func() {
		c.Tasks.Add(func() {
			ai, snap := reactable.NewBloomAttack(char, b, func(atk *combat.AttackInfo) {
				// atk.Abil += " (bountiful core)"
				// FIXME: some external code only match against AttackTagBloom. fix A4 if you uncomment this
				// atk.AttackTag = attacks.AttackTagBountifulCore
				atk.ICDTag = attacks.ICDTagBountifulCoreDamage
			})
			ap := combat.NewCircleHitOnTarget(b.Gadget, nil, 6.5)
			c.QueueAttackWithSnap(ai, snap, ap, 0)

			// self damage
			ai.Abil += reactions.SelfDamageSuffix
			ai.FlatDmg = 0.05 * ai.FlatDmg
			ap.SkipTargets[targets.TargettablePlayer] = false
			ap.SkipTargets[targets.TargettableEnemy] = true
			ap.SkipTargets[targets.TargettableGadget] = true
			c.QueueAttackWithSnap(ai, snap, ap, 0)
		}, 1)
	}
	b.Gadget.OnExpiry = explode
	b.Gadget.OnKill = explode

	return b
}

func (b *BountifulCore) Tick() {
	// this is needed since gadget tick
	b.Gadget.Tick()
}

func (b *BountifulCore) HandleAttack(atk *combat.AttackEvent) float64 {
	b.Core.Events.Emit(event.OnGadgetHit, b, atk)
	return 0
}
func (b *BountifulCore) Attack(*combat.AttackEvent, glog.Event) (float64, bool) { return 0, false }
func (b *BountifulCore) SetDirection(trg geometry.Point)                        {}
func (b *BountifulCore) SetDirectionToClosestEnemy()                            {}
func (b *BountifulCore) CalcTempDirection(trg geometry.Point) geometry.Point {
	return geometry.DefaultDirection()
}

