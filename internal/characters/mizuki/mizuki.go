package mizuki

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

func init() {
	core.RegisterCharFunc(keys.Mizuki, NewChar)
}

type char struct {
	*tmpl.Character
	dreamdrifterSrc int
	particleCount   int
	c4Count         int
	c2buff          []float64
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.SkillCon = 3
	c.BurstCon = 5

	c.particleCount = 0
	c.c4Count = 0

	w.Character = &c

	c.c2buff = make([]float64, attributes.EndStatType)

	return nil
}

func (c *char) Init() error {
	c.onExitField()
	c.swirlBuff()
	c.snackHandler("init")
	c.a1()
	c.a4()
	c.c1()
	c.c6()
	return nil
}

func (c *char) onExitField() {
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(_ ...interface{}) bool {
		if c.StatModIsActive(skillKey) {
			c.DeleteStatMod(skillKey)
		}
		return false
	}, "mizuki-exit")
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "dreamdrifter":
		if c.StatusIsActive(skillKey) {
			return 1, nil
		}
		return 0, nil
	default:
		return c.Character.Condition(fields)
	}
}

func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	if a == action.ActionSkill && c.StatusIsActive(skillKey) {
		return true, action.NoFailure
	}
	return c.Character.ActionReady(a, p)
}
