package sayu

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/hacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Sayu, NewChar)
	hacks.RegisterNOSpecialChar(keys.Sayu)
}

type char struct {
	*tmpl.Character
	eDuration           int
	eAbsorb             attributes.Element
	eAbsorbTag          attacks.ICDTag
	absorbCheckLocation combat.AttackPattern
	qTickRadius         float64
	c2Bonus             float64
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 80
	c.NormalHitNum = normalHitNum
	c.BurstCon = 3
	c.SkillCon = 5

	c.eDuration = -1
	c.eAbsorb = attributes.NoElement
	c.qTickRadius = 1
	c.c2Bonus = .0

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.a1()
	c.a4()
	c.rollAbsorb()
	if c.Base.Cons >= 2 {
		c.c2()
	}
	return nil
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 24
	}
	return c.Character.AnimationStartDelay(k)
}

