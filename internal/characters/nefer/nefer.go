package nefer

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Nefer, NewChar)
}

type char struct {
	*tmpl.Character
	a1count   float64
	c4buffkey bool
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.SkillCon = 3
	c.SetNumCharges(action.ActionSkill, 2)
	c.BurstCon = 5

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	// mark this character as a potential moonsign holder for team initialization
	c.AddStatus("moonsignKey", -1, false)
	c.InitLCallback()
	c.makeBurstBonus()
	c.a0()
	c.a1()
	if c.Base.Cons >= 4 {
		c.c4()
	}
	if c.Base.Cons >= 6 {
		c.c6()
	}
	return nil
}

func (c *char) ActionStam(a action.Action, p map[string]int) float64 {
	// If charging during Shadow Dance (skill active), stamina cost is 0 when Verdant Dew >=1.
	if a != action.ActionCharge || !c.StatusIsActive(skillKey) {
		return c.Character.ActionStam(a, p)
	}
	if c.Core.Player.Verdant.Count() >= 1 {
		return 0
	}
	return 25
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 11
	}
	return c.Character.AnimationStartDelay(k)
}

func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	return c.Character.ActionReady(a, p)
}

func (c *char) Condition(fields []string) (any, error) {
	return c.Character.Condition(fields)
}
