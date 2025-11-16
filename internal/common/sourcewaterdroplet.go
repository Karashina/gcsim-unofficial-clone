package common

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/gadget"
)

type SourcewaterDroplet struct {
	*gadget.Gadget
}

func NewSourcewaterDroplet(core *core.Core, pos geometry.Point, typ combat.GadgetTyp) *SourcewaterDroplet {
	p := &SourcewaterDroplet{}
	p.Gadget = gadget.New(core, pos, 1, typ)
	p.Gadget.Duration = 878
	core.Combat.AddGadget(p)
	return p
}

func (s *SourcewaterDroplet) HandleAttack(*combat.AttackEvent) float64 { return 0 }
func (s *SourcewaterDroplet) SetDirection(trg geometry.Point)          {}
func (s *SourcewaterDroplet) SetDirectionToClosestEnemy()              {}
func (s *SourcewaterDroplet) CalcTempDirection(trg geometry.Point) geometry.Point {
	return geometry.DefaultDirection()
}

func (s *SourcewaterDroplet) Type() targets.TargettableType                          { return targets.TargettableGadget }
func (s *SourcewaterDroplet) Attack(*combat.AttackEvent, glog.Event) (float64, bool) { return 0, false }

