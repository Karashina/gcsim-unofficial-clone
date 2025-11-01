package flins

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
	core.RegisterCharFunc(keys.Flins, NewChar)
}

type char struct {
	*tmpl.Character
	northlandCD int
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)
	c.SkillCon = 5
	c.BurstCon = 3

	c.EnergyMax = 80
	c.NormalHitNum = normalHitNum

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.AddStatus("moonsignKey", -1, false)
	c.InitLCallback()
	c.a0()
	c.a1()
	c.a4()
	c.c1()
	if c.Base.Cons >= 2 {
		c.c2()
	}
	if c.Base.Cons >= 4 {
		c.c4()
	}
	if c.Base.Cons >= 6 {
		c.c6()
	}
	return nil
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 11
	}
	return c.Character.AnimationStartDelay(k)
}

func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	if a == action.ActionSkill && c.StatusIsActive(skillKey) {
		if c.StatusIsActive(northlandCdKey) {
			return false, action.SkillCD // Fails if Northland Spearstorm is still on CD (This CD is unaffected by other effects such as c.CDReduction())
		}
		return true, action.NoFailure // Make Northland Spearstorm usable even on normal skill is in CD
	}
	if a == action.ActionBurst && c.StatusIsActive(northlandKey) {
		if !c.Core.Flags.IgnoreBurstEnergy && c.Energy < 30 {
			return false, action.InsufficientEnergy // Energy cost of Thunderous Symphony is 30
		}
		return true, action.NoFailure // Make Thunderous Symphony usable even on normal burst is in CD
	}
	return c.Character.ActionReady(a, p)
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "northlandup":
		return c.StatusIsActive(northlandCdKey), nil
	default:
		return c.Character.Condition(fields)
	}
}
