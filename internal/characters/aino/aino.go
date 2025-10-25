package aino

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Aino, NewChar)
}

type char struct {
	*tmpl.Character
	burstSrc      int
	c2IcdKey      string
	a4FlatDmgBuff float64
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)
	c.SkillCon = 5
	c.BurstCon = 3

	c.EnergyMax = 50
	c.NormalHitNum = normalHitNum

	c.c2IcdKey = "aino-c2-icd"

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.AddStatus("moonsignKey", -1, false)
	c.moonsignInitFunc()
	c.a1()
	c.a4()
	c.c1()
	c.c4()
	c.c6()
	return nil
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 11
	}
	return c.Character.AnimationStartDelay(k)
}

func (c *char) moonsignInitFunc() {
	count := 0
	for _, char := range c.Core.Player.Chars() {
		if char.StatusIsActive("moonsignKey") {
			count++
		}
	}
	switch count {
	case 1:
		c.MoonsignNascent = true
		c.MoonsignAscendant = false
	case 2, 3, 4:
		c.MoonsignAscendant = true
		c.MoonsignNascent = false
	default:
		c.MoonsignNascent = false
		c.MoonsignAscendant = false
	}
}
