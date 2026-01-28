package kachina

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/construct"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

type TurboTwirly struct {
	src    int
	expiry int
	c      *char
	dir    geometry.Point
	pos    geometry.Point
}

func (s *TurboTwirly) OnDestruct() {
}

func (s *TurboTwirly) Key() int                         { return s.src }
func (s *TurboTwirly) Type() construct.GeoConstructType { return construct.GeoConstructKachinaSkill }
func (s *TurboTwirly) Expiry() int                      { return s.expiry }
func (s *TurboTwirly) IsLimited() bool                  { return true }
func (s *TurboTwirly) Count() int                       { return 1 }
func (s *TurboTwirly) Direction() geometry.Point        { return s.dir }
func (s *TurboTwirly) Pos() geometry.Point              { return s.pos }
