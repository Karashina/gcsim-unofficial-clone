package skirk

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Skirk, NewChar)
}

type char struct {
	*tmpl.Character
	sevenPhaseFlashsrc  int
	isOnlyFrozenTeam    bool
	onSevenPhaseFlash   bool
	serpentsSubtlety    float64
	serpentsSubtletyMax float64
	voidrift            int
	deathsCrossing      []string
	a4BuffNA            float64
	a4BuffQ             float64
	burstbuffcount      int
	c6count             int
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 0
	c.NormalHitNum = normalHitNum
	c.BurstCon = 3
	c.SkillCon = 5

	// skill
	c.onSevenPhaseFlash = false
	c.serpentsSubtlety = 0
	c.serpentsSubtletyMax = 100

	//a0
	chars := c.Core.Player.Chars()
	count := make(map[attributes.Element]int)
	for _, this := range chars {
		count[this.Base.Element]++
	}
	c.isOnlyFrozenTeam = count[attributes.Hydro] > 0 && count[attributes.Hydro]+count[attributes.Cryo] == len(chars)
	// a4
	c.deathsCrossing = make([]string, 3)
	c.a4BuffNA = 1
	c.a4BuffQ = 1

	w.Character = &c
	return nil
}

func (c *char) Init() error {
	for _, char := range c.Core.Player.Chars() {
		char.SetTag(keys.SkirkPassive, 1)
	}

	c.a1(true)
	c.a4()
	c.c4()
	c.onExitField()
	return nil
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	switch k {
	case model.AnimationXingqiuN0StartDelay:
		return 14
	case model.AnimationYelanN0StartDelay:
		return 4
	default:
		return c.Character.AnimationStartDelay(k)
	}
}

func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	if a == action.ActionBurst && !c.onSevenPhaseFlash {
		if c.serpentsSubtlety >= 50 {
			return true, action.NoFailure
		} else {
			return false, action.InsufficientEnergy
		}
	}
	return c.Character.ActionReady(a, p)
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "sevenphase":
		return c.onSevenPhaseFlash, nil
	case "voidrift":
		return c.voidrift, nil
	default:
		return c.Character.Condition(fields)
	}
}

func (c *char) NextQueueItemIsValid(k keys.Char, a action.Action, p map[string]int) error {
	if a == action.ActionCharge {
		return nil
	}
	return c.Character.NextQueueItemIsValid(k, a, p)
}
