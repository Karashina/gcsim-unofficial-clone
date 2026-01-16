package lauma

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
	core.RegisterCharFunc(keys.Lauma, NewChar)
}

type char struct {
	*tmpl.Character
	skillSrc    int
	burstLBBuff float64
	verdantDew  int
	moonSong    int
	paleHymn    int
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
	// mark this character as a potential moonsign holder for team initialization
	c.AddStatus("moonsignKey", -1, false)
	c.setupPaleHymnEffects()
	c.a0()
	c.a4()                // initialize A4 AddAttackMod
	c.c6Init()            // initialize C6 Ascendant Elevation bonus
	c.verdantDewCheck()   // initialize Verdant Dew monitoring
	c.applyResReduction() // initialize RES reduction monitoring
	return nil
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 11
	}
	return c.Character.AnimationStartDelay(k)
}

// verdantDewCheck subscribes to Bloom events and grants Verdant Dew to the party.
// Verdant Dew is capped at 3 and gains are queued after the charging period.
func (c *char) verdantDewCheck() {
	c.Core.Events.Subscribe(event.OnBloom, func(args ...interface{}) bool {
		if !c.StatusIsActive("LB-Key") {
			return false
		}
		duradd := c.StatusDuration("dewchargingkey")
		c.AddStatus("dewchargingkey", 150, true) // 2.5s charging period
		c.QueueCharTask(func() {
			if c.verdantDew < 3 {
				c.verdantDew++
			}
		}, 150+duradd)
		return true
	}, "lauma-verdant-dew")
}
