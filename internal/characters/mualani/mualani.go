package mualani

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/internal/template/nightsoul"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Mualani, NewChar)
}

type char struct {
	*tmpl.Character
	nightsoulState *nightsoul.State
	nightsoulSrc   int
	momentumStacks int
	momentumSrc    int
	a4Stacks       int
	c1Done         bool

	a1Count int
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
	c.nightsoulState.MaxPoints = 60

	return nil
}

func (c *char) Init() error {
	c.a4()

	c.c4()

	c.SetNumCharges(action.ActionAttack, 1)
	c.onExitField()
	c.surfingTick()
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
	case "momentum":
		return c.momentumStacks, nil
	default:
		return c.Character.Condition(fields)
	}
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if c.nightsoulState.HasBlessing() {
		if c.momentumStacks >= 3 {
			switch k {
			case model.AnimationXingqiuN0StartDelay:
				return 44
			default:
				return 37
			}
		}
		switch k {
		case model.AnimationXingqiuN0StartDelay:
			return 11
		default:
			return 9
		}
	}
	switch k {
	case model.AnimationXingqiuN0StartDelay:
		return 11
	default:
		return 11
	}
}

func (c *char) ActionStam(a action.Action, p map[string]int) float64 {
	if c.nightsoulState.HasBlessing() {
		return 0
	}
	return c.Character.ActionStam(a, p)
}

func (c *char) NextQueueItemIsValid(k keys.Char, a action.Action, p map[string]int) error {
	if c.nightsoulState.HasBlessing() {
		// cannot CA in nightsoul blessing
		if a == action.ActionCharge {
			return player.ErrInvalidChargeAction
		}
	}

	return c.Character.NextQueueItemIsValid(k, a, p)
}

func (c *char) onExitField() {
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(_ ...interface{}) bool {
		if c.nightsoulState.HasBlessing() {
			c.cancelNightsoul()
		}
		return false
	}, "mualani-exit")
}
