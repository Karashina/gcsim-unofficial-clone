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
	skillSrc    int
	burstLBBuff float64
	verdantDew  int
	moonSong    int
	paleHymn    int
	a4crval     float64
	a4cdval     float64
	c6mult      float64
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
	c.a4crval = 0
	c.a4cdval = 0
	c.moonsignInitFunc()
	c.setupPaleHymnEffects()
	c.a0()
	c.a4()                               // Initialize A4 AddAttackMod
	c.c6mult = c.c6AscendantMultiplier() // Apply C6 Ascendant multiplier if applicable
	c.verdantDewCheck()                  // Initialize Verdant Dew monitoring
	c.applyResReduction()                // Initialize RES reduction monitoring
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
func (c *char) verdantDewCheck() {
	c.Core.Events.Subscribe(event.OnBloom, func(args ...interface{}) bool {
		if c.StatusIsActive("LB-Key") {
			duradd := c.StatusDuration("dewchargingkey")
			c.AddStatus("dewchargingkey", 150, true) // 2.5s charging period
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

func (c *char) moonsignInitFunc() {
	count := 0
	for _, char := range c.Core.Player.Chars() {
		if char.StatusIsActive("moonsignKey") {
			count++
		}
	}
	switch count {
	case 1:
		c.MoonsignNascent = true // Moonsign: Nascent Gleam
		c.MoonsignAscendant = false
	case 2, 3, 4:
		c.MoonsignAscendant = true // Moonsign: Ascendant Gleam
		c.MoonsignNascent = false
	default:
		c.MoonsignNascent = false
		c.MoonsignAscendant = false
	}
}
