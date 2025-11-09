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
	burstSrc int
	c2IcdKey string
	// whether Aino's Moonsign is Ascendant (affects burst interval and bonuses)
	MoonsignAscendant bool
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
	// mark this character as a potential moonsign holder for team initialization
	c.AddStatus("moonsignKey", -1, false)
	c.c1()
	c.c2()
	c.c4()
	c.c6()
	return nil
}

func (c *char) AnimationStartDelay(k info.AnimationDelayKey) int {
	if k == info.AnimationDelayKey(info.AnimationXingqiuN0StartDelay) {
		return 11
	}
	return c.Character.AnimationStartDelay(k)
}
