package mavuika

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/internal/template/nightsoul"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Mavuika, NewChar)
}

type char struct {
	*tmpl.Character
	nightsoulState    *nightsoul.State
	nightsoulSrc      int
	normalBikeCounter int
	fightingspirit    float64
	maxfightingspirit float64
	consumedspirit    float64
	a4buff            []float64
	c2trg             []combat.Enemy
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 0
	c.NormalHitNum = normalHitNum
	c.SkillCon = 5
	c.BurstCon = 3
	c.HasArkhe = false

	w.Character = &c

	c.nightsoulState = nightsoul.New(s, w)
	c.nightsoulState.MaxPoints = 80
	if c.Base.Cons >= 1 {
		c.nightsoulState.MaxPoints = 120
	}

	c.maxfightingspirit = 200
	c.fightingspirit = c.maxfightingspirit

	c.consumedspirit = 0

	return nil
}

func (c *char) Init() error {
	c.a4buff = make([]float64, attributes.EndStatType)
	c.a1()
	c.c2()
	c.GainFightingSpirit()
	c.onExitField()
	return nil
}

func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	if a == action.ActionSkill && c.nightsoulState.HasBlessing() {
		return true, action.NoFailure
	}
	if a == action.ActionBurst {
		if !c.StatusIsActive(BurstCDKey) {
			if c.fightingspirit >= 100 {
				return true, action.NoFailure
			} else {
				return false, action.InsufficientEnergy
			}
		} else {
			return false, action.BurstCD
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

func (c *char) ActionStam(a action.Action, p map[string]int) float64 {
	if c.nightsoulState.HasBlessing() {
		return 0
	}
	return c.Character.ActionStam(a, p)
}

func (c *char) onExitField() {
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(_ ...interface{}) bool {
		if c.StatusIsActive(bikeKey) {
			c.DeleteStatus(bikeKey)
		}
		return false
	}, "mavuika-exit")
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	switch k {
	case model.AnimationXingqiuN0StartDelay:
		return 19
	case model.AnimationYelanN0StartDelay:
		return 19
	default:
		return c.Character.AnimationStartDelay(k)
	}
}

func (c *char) AdvanceNormalIndex() {
	if c.nightsoulState.HasBlessing() {
		c.normalBikeCounter++
		if c.normalBikeCounter == skillHitNum {
			c.normalBikeCounter = 0
		}
		return
	}
	c.NormalCounter++
	if c.NormalCounter == c.NormalHitNum {
		c.NormalCounter = 0
	}
}
