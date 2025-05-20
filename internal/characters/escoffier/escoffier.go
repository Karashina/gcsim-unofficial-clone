package escoffier

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Escoffier, NewChar)
}

type char struct {
	*tmpl.Character
	cookingMekSrc int
	a1Src         int
	c2count       int
	c4count       int
	c6count       int
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.BurstCon = 5
	c.SkillCon = 3
	c.HasArkhe = true

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.a4()
	c.c2()
	c.c6()
	return nil
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	switch k {
	case model.AnimationXingqiuN0StartDelay:
		return 15
	case model.AnimationYelanN0StartDelay:
		return 6
	default:
		return c.Character.AnimationStartDelay(k)
	}
}
