package kinich

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Kinich, NewChar)
}

type char struct {
	*tmpl.Character
	skilltravel int
	skillhold   int
	isSSCHeld   bool
	sscradius   float64
	a4stacks    int
	a4buff      float64
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.CharZone = info.ZoneNatlan
	c.EnergyMax = 70
	c.NormalHitNum = normalHitNum
	c.SkillCon = 3
	c.BurstCon = 5

	c.NightsoulPointMax = 20
	c.NightsoulPoint = 0
	c.HasNightsoul = true
	c.OnNightsoul = false

	c.a4stacks = 0
	c.a4buff = 0

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.onExitField()
	c.skillendcheck()
	c.a1mark()
	c.a1regen()
	c.a4()
	c.c1()
	c.c2()
	c.c4()
	return nil
}

func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	if a == action.ActionSkill && c.StatusIsActive(skillKey) && c.NightsoulPoint >= c.NightsoulPointMax {
		return true, action.NoFailure
	}
	return c.Character.ActionReady(a, p)
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "nightsoul-points":
		if c.NightsoulPoint <= 0 {
			return 0, nil
		}
		return c.NightsoulPoint, nil
	case "onnightsoul":
		if !c.OnNightsoul {
			return 0, nil
		}
		return 1, nil
	default:
		return c.Character.Condition(fields)
	}
}

func (c *char) onExitField() {
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(_ ...interface{}) bool {
		if c.StatModIsActive(skillKey) {
			c.skillEndRoutine()
		}
		return false
	}, "kinich-exit")
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	switch k {
	case model.AnimationXingqiuN0StartDelay:
		if c.StatusIsActive(skillKey) {
			return 12
		}
		return 0
	default:
		return c.Character.AnimationStartDelay(k)
	}
}
