package varesa

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/internal/template/nightsoul"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/core/stacks"
)

func init() {
	core.RegisterCharFunc(keys.Varesa, NewChar)
}

const (
	fieryPassionKey = "fiery-passion"
	freeskillkey    = "varesa-free-skill"
	apexDriveKey    = "varesa-apex-drive"
)

type char struct {
	*tmpl.Character
	nightsoulState    *nightsoul.State
	particleGenerated bool
	fastCharge        bool
	a4Stacks          *stacks.MultipleRefreshNoRemove
	c4buff            float64
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 70
	c.NormalHitNum = normalHitNum
	c.NormalCon = 5
	c.BurstCon = 3
	c.HasArkhe = false

	w.Character = &c

	c.nightsoulState = nightsoul.New(s, w)
	c.nightsoulState.MaxPoints = 40

	c.SetNumCharges(action.ActionSkill, 2)
	return nil
}

func (c *char) Init() error {
	c.Nightsoulckeck()
	c.a4()
	c.c6()
	return nil
}

func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	if a == action.ActionSkill && c.StatusIsActive(freeskillkey) {
		return true, action.NoFailure
	}
	if a == action.ActionBurst && c.StatusIsActive(apexDriveKey) && c.Energy >= kablamCost && c.Cooldown(action.ActionBurst) <= 0 {
		return true, action.NoFailure
	}
	return c.Character.ActionReady(a, p)
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "nightsoul":
		return c.nightsoulState.Condition(fields)
	case "fiery":
		if c.StatusIsActive(fieryPassionKey) {
			return 1, nil
		} else {
			return 0, nil
		}
	default:
		return c.Character.Condition(fields)
	}
}

func (c *char) ActionStam(a action.Action, p map[string]int) float64 {
	if a == action.ActionCharge && c.StatusIsActive(skillKey) {
		return 0
	}
	return c.Character.ActionStam(a, p)
}

func (c *char) Nightsoulckeck() {
	c.Core.Events.Subscribe(event.OnNightsoulGenerate, func(args ...interface{}) bool {
		idx := args[0].(int)
		if c.Index != idx {
			return false
		}
		if c.nightsoulState.Points() >= 40 {
			if !c.nightsoulState.HasBlessing() {
				c.nightsoulState.EnterBlessing(40)
			}
			c.AddStatus(fieryPassionKey, 15*60, false)
			c.AddStatus(freeskillkey, 15*60, true)
		}
		return false
	}, "varesa-fierypassion-check")

	c.Core.Events.Subscribe(event.OnTick, func(args ...interface{}) bool {
		if !c.nightsoulState.HasBlessing() || c.StatusIsActive(fieryPassionKey) {
			return false
		}
		if !c.StatusIsActive(fieryPassionKey) {
			c.nightsoulState.ExitBlessing()
		}
		return false
	}, "varesa-nightsoul-check")

	c.Core.Events.Subscribe(event.OnActionExec, func(args ...interface{}) bool {
		idx := args[0].(int)
		if c.Index != idx {
			return false
		}
		act := args[1].(action.Action)
		if act == action.ActionSkill {
			c.DeleteStatus(apexDriveKey)
			return false
		}
		if act != action.ActionHighPlunge && act != action.ActionLowPlunge {
			return false
		}
		if c.Index != c.Core.Player.ActiveChar().Index {
			return false
		}
		if !c.StatusIsActive(fieryPassionKey) {
			if c.Base.Cons >= 2 {
				c.AddStatus(apexDriveKey, 5*60, true)
			}
			return false
		}
		c.AddStatus(apexDriveKey, 5*60, true)
		return false
	}, "varesa-apexdrive-check")
}
