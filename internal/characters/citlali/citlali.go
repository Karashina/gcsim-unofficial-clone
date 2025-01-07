package citlali

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/internal/template/nightsoul"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Citlali, NewChar)
}

type char struct {
	*tmpl.Character
	nightsoulState *nightsoul.State
	nightsoulSrc   int
	c6count        float64
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.SkillCon = 3
	c.BurstCon = 5
	c.HasArkhe = false

	w.Character = &c

	c.nightsoulState = nightsoul.New(s, w)
	c.nightsoulState.MaxPoints = 100

	return nil
}

func (c *char) Init() error {
	c.SkillChecks()
	c.a1()
	c.a4()
	c.c1()
	c.c2()
	c.c6()
	return nil
}

func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	if a == action.ActionAttack && c.nightsoulState.HasBlessing() {
		if c.AvailableCDCharge[a] <= 0 {
			// TODO: Implement AttackCD warning
			return false, action.CharacterDeceased
		}
	}

	return c.Character.ActionReady(a, p)
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "nightsoul":
		return c.nightsoulState.Condition(fields)
	default:
		return c.Character.Condition(fields)
	}
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 4
	}
	return c.Character.AnimationStartDelay(k)
}

func (c *char) ActionStam(a action.Action, p map[string]int) float64 {
	if c.nightsoulState.HasBlessing() {
		return 0
	}
	return c.Character.ActionStam(a, p)
}
