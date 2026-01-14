package sethos

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Sethos, NewChar)
}

type char struct {
	*tmpl.Character
	lastSkillFrame int
	a4Count        int
	c4Buff         []float64
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}

	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.BurstCon = 5
	c.NormalCon = 3

	c.lastSkillFrame = -1

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.skillRefundHook()
	c.a4()
	c.c1()
	c.c2()
	c.c4()
	c.onExitField()
	return nil
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	switch k {
	case model.AnimationXingqiuN0StartDelay:
		return 5
	case model.AnimationYelanN0StartDelay:
		return 4
	default:
		return c.Character.AnimationStartDelay(k)
	}
}

func (c *char) onExitField() {
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(_ ...interface{}) bool {
		if c.StatusIsActive(burstBuffKey) {
			c.DeleteStatus(burstBuffKey)
		}
		return false
	}, "sethos-exit")
}

