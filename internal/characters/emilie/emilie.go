package emilie

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Emilie, NewChar)
}

type char struct {
	*tmpl.Character

	caseTravel   int
	lumidouceSrc int
	lumidoucePos geometry.Point

	prevLumidouceLvl  int
	burstMarkDuration int

	c6Scents int
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 50
	c.NormalHitNum = normalHitNum
	c.BurstCon = 5
	c.SkillCon = 3
	c.HasArkhe = true

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.a4()
	c.c1()

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
