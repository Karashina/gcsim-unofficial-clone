package yaemiko

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

const (
	yaeTotemCount  = "totems"
	yaeTotemStatus = "yae_oldest_totem_expiry"
)

func init() {
	core.RegisterCharFunc(keys.YaeMiko, NewChar)
}

type char struct {
	*tmpl.Character
	kitsuneDetectionRadius float64
	kitsunes               []*kitsune
	c4buff                 []float64
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 90
	c.NormalHitNum = normalHitNum
	c.BurstCon = 5
	c.SkillCon = 3

	c.SetNumCharges(action.ActionSkill, 3)

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.a4()
	if c.Base.Cons >= 2 {
		c.kitsuneDetectionRadius = 20
	} else {
		c.kitsuneDetectionRadius = 12.5
	}
	if c.Base.Cons >= 4 {
		c.c4buff = make([]float64, attributes.EndStatType)
		c.c4buff[attributes.ElectroP] = .20
	}
	return nil
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 7
	}
	return c.Character.AnimationStartDelay(k)
}
