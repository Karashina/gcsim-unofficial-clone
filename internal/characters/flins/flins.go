package flins

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
	// mark this character as a potential moonsign holder for team initialization
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
	// Allow special skill/burst variants when their custom statuses are active.
	if a == action.ActionSkill && c.StatusIsActive(skillKey) {
		if c.StatusIsActive(northlandCdKey) {
			// Northland Spearstorm is on its own CD and cannot be used yet.
			return false, action.SkillCD
		}
		// Skill variant allowed even if normal skill CD is active.
		return true, action.NoFailure
	}

	if a == action.ActionBurst && c.StatusIsActive(northlandKey) {
		if !c.Core.Flags.IgnoreBurstEnergy && c.Energy < 30 {
			// Thunderous Symphony requires 30 energy unless IgnoreBurstEnergy is set.
			return false, action.InsufficientEnergy
		}
		// Burst variant allowed even if normal burst CD is active.
		return true, action.NoFailure
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
