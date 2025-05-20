package ifa

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/internal/template/nightsoul"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Ifa, NewChar)
}

type char struct {
	*tmpl.Character
	charnightsoulpoints []float64
	nightsoulState      *nightsoul.State
	nightsoulSrc        int
	skillParticleICD    bool
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)
	c.nightsoulState = nightsoul.New(c.Core, c.CharWrapper)
	c.nightsoulState.MaxPoints = 80
	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.SkillCon = 3
	c.BurstCon = 5
	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.charnightsoulpoints = make([]float64, len(c.Core.Player.Chars()))
	c.checkCurrentNightsoulPoint()
	c.a1()
	c.a4()
	return nil
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "nightsoul":
		return c.nightsoulState.Condition(fields)
	default:
		return c.Character.Condition(fields)
	}
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if c.nightsoulState.HasBlessing() {
		switch k {
		case model.AnimationXingqiuN0StartDelay:
			return 5
		case model.AnimationYelanN0StartDelay:
			return 0
		default:
			return c.Character.AnimationStartDelay(k)
		}
	}
	switch k {
	case model.AnimationXingqiuN0StartDelay:
		return 12
	case model.AnimationYelanN0StartDelay:
		return 5
	default:
		return c.Character.AnimationStartDelay(k)
	}
}
