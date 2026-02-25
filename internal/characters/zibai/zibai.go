package zibai

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
	core.RegisterCharFunc(keys.Zibai, NewChar)
}

type char struct {
	*tmpl.Character
	// Lunar Phase Shift state
	lunarPhaseShiftActive bool
	lunarPhaseShiftSrc    int
	phaseShiftRadiance    int // Current radiance points
	maxPhaseShiftRadiance int // Default 300
	spiritSteedUsages     int // Current usages per mode
	maxSpiritSteedUsages  int // Default 3, C1 increases to 5
	savedNormalCounter    int // Saved normal counter for C4
	// Constellation tracking
	c1FirstStride     bool // C1: first stride bonus
	c4ScattermoonUsed bool // C4: next N4 does 250% damage

}

const (
	skillKey                 = "zibai-lunar-phase-shift"
	selenicDescentKey        = "zibai-selenic-descent"
	spiritSteedCDKey         = "zibai-spirit-steed-cd"
	particleICDKey           = "zibai-particle-icd"
	radianceNormalICDKey     = "zibai-radiance-na-icd"
	radianceLCrsICDKey       = "zibai-radiance-lcrs-icd"
	lunarPhaseShiftDuration  = 990 // 16.5 seconds
	spiritSteedRadianceCost  = 70  // Cost to use Spirit Steed's Stride
	normalPhaseShiftRadiance = 100 // Default max radiance (spec: 最大100まで)
	c6ElevationBuffKey       = "zibai-c6-elevation"
	radianceTickInterval     = 6
	radianceTickGain         = 1
	radianceNormalGain       = 5
	radianceNormalICD        = 30
	radianceLCrsGain         = 35
	radianceLCrsICD          = 4 * 60
)

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.BurstCon = 5
	c.SkillCon = 3
	// Initialize state
	c.lunarPhaseShiftActive = false
	c.phaseShiftRadiance = 0
	c.maxPhaseShiftRadiance = normalPhaseShiftRadiance
	c.maxSpiritSteedUsages = 4
	c.spiritSteedUsages = 0
	c.c1FirstStride = false
	c.c4ScattermoonUsed = false

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	// Initialize Moonsign passive (grants LCrs-Key to party)
	c.a0Init()
	// Initialize A1 passive
	c.a1Init()
	// Initialize A4 passive
	if c.Base.Ascension >= 4 {
		c.a4Init()
	}
	// Initialize constellations
	if c.Base.Cons >= 1 {
		c.c1Init()
	}
	if c.Base.Cons >= 2 {
		c.c2Init()
	}
	if c.Base.Cons >= 4 {
		c.c4Init()
	}
	if c.Base.Cons >= 6 {
		c.c6Init()
	}

	c.initRadianceHandlers()

	return nil
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 10
	}
	return c.Character.AnimationStartDelay(k)
}

// Condition allows querying character state from GCSL
func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "lunar-phase-shift":
		if c.lunarPhaseShiftActive {
			return 1, nil
		}
		return 0, nil
	case "phase-shift-radiance":
		return c.phaseShiftRadiance, nil
	case "spirit-steed-usages":
		return c.spiritSteedUsages, nil
	case "moonsign-ascendant":
		if c.MoonsignAscendant {
			return 1, nil
		}
		return 0, nil
	}
	return c.Character.Condition(fields)
}

// ActionReady checks if the Spirit Steed's Stride can be used
func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	// Spirit Steed's Stride special check
	if a == action.ActionSkill && c.lunarPhaseShiftActive {
		// Check if can recast (has radiance and usages)
		if c.phaseShiftRadiance < spiritSteedRadianceCost {
			return false, action.InsufficientEnergy
		}
		if c.spiritSteedUsages >= c.maxSpiritSteedUsages {
			return false, action.SkillCD
		}
		return true, action.NoFailure
	}
	return c.Character.ActionReady(a, p)
}

// isMoonsignAscendant checks if Moonsign is Ascendant Gleam
func (c *char) isMoonsignAscendant() bool {
	return c.MoonsignAscendant
}

// addPhaseShiftRadiance adds radiance points, capped at max
func (c *char) addPhaseShiftRadiance(amount int) {
	// C6: 50% increased gain rate
	if c.Base.Cons >= 6 {
		amount = int(float64(amount) * 1.5)
	}
	c.phaseShiftRadiance += amount
	if c.phaseShiftRadiance > c.maxPhaseShiftRadiance {
		c.phaseShiftRadiance = c.maxPhaseShiftRadiance
	}
}
