package razor

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
	core.RegisterCharFunc(keys.Razor, NewChar)
}

type char struct {
	*tmpl.Character
	sigils  int
	a4Bonus []float64
	c1bonus []float64
	c2bonus []float64
	// Hexerei mode (default true unless nohex=1)
	isHexerei bool
}

func NewChar(s *core.Core, w *character.CharWrapper, p info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 80
	c.BurstCon = 3
	c.SkillCon = 5
	c.NormalHitNum = normalHitNum

	// Default is Hexerei character unless nohex=1 is specified
	c.isHexerei = true
	if nohex, ok := p.Params["nohex"]; ok && nohex == 1 {
		c.isHexerei = false
	}

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	// burst
	c.onSwapClearBurst()

	c.a4()

	// make sure to use the same key everywhere so that these passives don't stack
	c.Core.Player.AddStamPercentMod("utility-dash", -1, func(a action.Action) (float64, bool) {
		if a == action.ActionDash && c.CurrentHPRatio() > 0 {
			return -0.2, false
		}
		return 0, false
	})
	if c.Base.Cons >= 1 {
		c.c1()
	}
	if c.Base.Cons >= 2 {
		c.c2()
	}

	return nil
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 18
	}
	return c.Character.AnimationStartDelay(k)
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "hexerei":
		return c.isHexerei, nil
	default:
		return c.Character.Condition(fields)
	}
}
