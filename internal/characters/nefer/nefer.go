package nefer

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/model"
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
	c.AddStatus("moonsignKey", -1, false)
	c.moonsignInitFunc()
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

func (c *char) moonsignInitFunc() {
	count := 0
	for _, ch := range c.Core.Player.Chars() {
		if ch.StatusIsActive("moonsignKey") {
			count++
		}
	}
	c.MoonsignNascent, c.MoonsignAscendant = false, false
	switch count {
	case 1:
		c.MoonsignNascent = true
	case 2, 3, 4:
		c.MoonsignAscendant = true
	}
}

func (c *char) ActionStam(a action.Action, p map[string]int) float64 {
	if a == action.ActionCharge && c.StatusIsActive(skillKey) {
		if c.Core.Player.Verdant.Count() >= 1 {
			return 0
		} else {
			return 25
		}
	}
	return c.Character.ActionStam(a, p)
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
