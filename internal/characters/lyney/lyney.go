package lyney

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
	core.RegisterCharFunc(keys.Lyney, NewChar)
}

type char struct {
	*tmpl.Character
	hats              []*GrinMalkinHat
	maxHatCount       int
	propSurplusStacks int
	pyrotechnicTravel int
	c2Src             int
	c2Stacks          int
}

func NewChar(s *core.Core, w *character.CharWrapper, p info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.BurstCon = 5
	c.NormalCon = 3
	c.HasArkhe = true

	c2Stacks, ok := p.Params["c2_stacks"]
	if !ok {
		c.c2Stacks = 0
	}
	if c2Stacks < 0 {
		c.c2Stacks = 0
	}
	if c2Stacks > 3 {
		c.c2Stacks = 3
	}

	pyrotechnicTravel, ok := p.Params["pyrotechnic_travel"]
	if ok {
		c.pyrotechnicTravel = pyrotechnicTravel
	} else {
		c.pyrotechnicTravel = 36
	}

	c.maxHatCount = 1
	c.c1()

	c.hats = make([]*GrinMalkinHat, 0, c.maxHatCount)

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.a4()

	c.c2Setup()

	return nil
}

func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	// check if it is possible to use E (in Q)
	if a == action.ActionSkill && c.StatusIsActive(burstKey) {
		return true, action.NoFailure
	}
	return c.Character.ActionReady(a, p)
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "propSurplusStacks":
		return c.propSurplusStacks, nil
	default:
		return c.Character.Condition(fields)
	}
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	switch k {
	case model.AnimationXingqiuN0StartDelay:
		return 12
	case model.AnimationYelanN0StartDelay:
		return 5
	default:
		return c.Character.AnimationStartDelay(k)
	}
}
