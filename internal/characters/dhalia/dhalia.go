package dhalia

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Dhalia, NewChar)
}

type char struct {
	*tmpl.Character
	favoniusfavor int
	nacount       int
	burstdur      int
	shieldExpiry  int
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.BurstCon = 5
	c.SkillCon = 3

	c.burstdur = 12 * 60
	if c.Base.Cons >= 4 {
		c.burstdur = 15 * 60
	}

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.generateFavonianFavor()
	c.regenShirld()
	return nil
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	switch k {
	case model.AnimationXingqiuN0StartDelay:
		return 13
	case model.AnimationYelanN0StartDelay:
		return 6
	default:
		return c.Character.AnimationStartDelay(k)
	}
}

