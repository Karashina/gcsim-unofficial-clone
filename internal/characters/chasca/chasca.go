package chasca

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/internal/template/nightsoul"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Chasca, NewChar)
}

type char struct {
	*tmpl.Character
	nightsoulState *nightsoul.State
	nightsoulSrc   int
	ElementSlot    []attributes.Element
	Shells         []attributes.Element
	typeCount      int
	anemoCount     int
	anemoremaining int
	a1Prob         float64
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
	c.nightsoulState.MaxPoints = 80

	return nil
}

func (c *char) Init() error {
	c.ElementSlot = make([]attributes.Element, 3)
	c.Shells = make([]attributes.Element, 6)
	c.onExitField()
	c.CheckShellElement()
	c.a1()
	c.a4()
	c.c6CDbuff()
	return nil
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

func (c *char) onExitField() {
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(_ ...interface{}) bool {
		if c.nightsoulState.HasBlessing() {
			c.cancelNightsoul()
		}
		return false
	}, "chasca-exit")
}

func (c *char) CheckShellElement() {
	chars := c.Core.Player.Chars()
	i := 0
	c.typeCount = 0
	pyroCount := 0
	hydroCount := 0
	electroCount := 0
	cryoCount := 0
	for _, this := range chars {
		if this.Index == c.Index {
			continue
		}
		switch this.Base.Element {
		case attributes.Pyro:
			c.ElementSlot[i] = attributes.Pyro
			if pyroCount == 0 {
				c.typeCount++
			}
			pyroCount++
		case attributes.Hydro:
			c.ElementSlot[i] = attributes.Hydro
			if hydroCount == 0 {
				c.typeCount++
			}
			hydroCount++
		case attributes.Cryo:
			c.ElementSlot[i] = attributes.Cryo
			if cryoCount == 0 {
				c.typeCount++
			}
			cryoCount++
		case attributes.Electro:
			c.ElementSlot[i] = attributes.Electro
			if electroCount == 0 {
				c.typeCount++
			}
			electroCount++
		default:
			c.anemoCount++
		}
		i++
	}
	c.Core.Log.NewEvent("Chasca Shells Init", glog.LogCharacterEvent, c.Index).
		Write("Slot-0", c.ElementSlot[0]).
		Write("Slot-1", c.ElementSlot[1]).
		Write("Slot-2", c.ElementSlot[2])
}
