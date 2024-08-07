package emilie

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Emilie, NewChar)
}

type char struct {
	*tmpl.Character
	// field use for calculating oz damage
	L1ai            combat.AttackInfo
	L2ai            combat.AttackInfo
	LCPos           geometry.Point
	LCSnapshot      combat.AttackEvent
	LCLevel         int
	LCSource        int  // keep tracks of source for resets
	LCActive        bool // purely used for gscl conditional purposes
	LCActiveUntil   int  // used for LC ticks
	LCTickSrc       int  // used for LC recast
	LCTravel        int
	Scents          int
	burstLCSpawnSrc int // prevent double LC spawn from burst
	c6Count         int
}

func NewChar(s *core.Core, w *character.CharWrapper, p info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 50
	c.NormalHitNum = normalHitNum
	c.SkillCon = 3
	c.BurstCon = 5

	c.LCSource = -1
	c.LCActive = false
	c.LCActiveUntil = -1
	c.LCTickSrc = -1
	c.LCLevel = 0
	c.Scents = 0

	c.LCTravel = 10
	travel, ok := p.Params["lc_travel"]
	if ok {
		c.LCTravel = travel
	}

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.ScentsCheck()
	c.a4()
	c.a0()
	c.c1()
	c.c1dendro()
	c.c1burn()
	c.c2()
	return nil
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "lc":
		return c.LCActive, nil
	case "lc-source":
		return c.LCSource, nil
	case "lc-duration":
		duration := c.LCActiveUntil - c.Core.F
		if duration < 0 {
			duration = 0
		}
		return duration, nil
	default:
		return c.Character.Condition(fields)
	}
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 9
	}
	return c.Character.AnimationStartDelay(k)
}
