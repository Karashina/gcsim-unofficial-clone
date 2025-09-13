package lauma

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Lauma, NewChar)
}

type char struct {
	*tmpl.Character
	skillSrc          int
	burstLBBuff       float64
	moonsignNascent   bool
	moonsignAscendant bool
	verdantDew        int
	moonSong          int
	paleHymn          int
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)
	c.SkillCon = 5
	c.BurstCon = 3

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.AddStatus("moonsignKey", -1, false)
	c.moonsignInitFunc()
	c.a0()
	return nil
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 11
	}
	return c.Character.AnimationStartDelay(k)
}

// Verdant Dew
// When party members trigger Lunar-Bloom reactions, the party will receive such a resource.
// Your party will receive 1 Verdant Dew every 2.5s for 2.5s after triggering a Lunar-Bloom reaction.
// Your party can have up to 3 Verdant Dew.
// If you trigger another Lunar-Bloom reaction during this time, the duration of this effect will be reset.
func (c *char) verdantDewCheck(k model.AnimationDelayKey) {
	c.Core.Events.Subscribe(event.OnBloom, func(args ...interface{}) bool {
		if c.StatusIsActive("LB-Key") {
			c.AddStatus("dewchargingkey", 150, true)
			duradd := 0
			duradd = c.StatusDuration("dewchargingkey")
			c.QueueCharTask(func() {
				if c.verdantDew < 3 {
					c.verdantDew++
				}
			}, 150+duradd)
			return true
		}
		return false
	}, "lauma-verdant-dew")
}
