package columbina

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
	core.RegisterCharFunc(keys.Columbina, NewChar)
}

type char struct {
	*tmpl.Character
	// Gravity system for Elemental Skill
	gravity           int
	gravityLC         int // Gravity from Lunar-Charged
	gravityLB         int // Gravity from Lunar-Bloom
	gravityLCrs       int // Gravity from Lunar-Crystallize
	gravityRippleSrc  int
	gravityRippleExp  int
	lunarDomainActive bool
	lunarDomainSrc    int
	lunacyStacks      int
	lunacySrc         int
	moonridgeDew      int
	moonridgeICD      int
	c4ICD             int
	c4DominantType    string // Track the dominant type for C4 bonus
	// Gravity Accumulation state
	activeGravityType string
	// A4 tracking
	a4LCSrc     int
	a4LCCount   int
	a4LBSrc     int
	a4LCrsSrc   int
	a4LCrsCount int
}

const (
	skillKey         = "columbina-skill"
	gravityRippleKey = "columbina-gravity-ripple"
	newMoonOmenKey   = "columbina-new-moon-omen"
	particleICDKey   = "columbina-particle-icd"
)

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.SkillCon = 3
	c.BurstCon = 5

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum

	// Initialize state
	c.gravityRippleSrc = -1
	c.lunacyStacks = 0
	c.moonridgeDew = 0
	c.moonridgeICD = 0
	c.c4ICD = 0

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	// Mark as moonsign holder and LCrs-Key holder for Lunar-Crystallize
	c.AddStatus("moonsignKey", -1, false)
	c.AddStatus("lcrs-key", -1, false)

	// Initialize passives
	c.a0Init()
	c.a1Init()
	// Initialize constellations
	if c.Base.Cons >= 1 {
		c.c1Init()
	}
	if c.Base.Cons >= 3 {
		c.c3c5Init()
	}
	if c.Base.Cons >= 6 {
		c.c6Init()
	}

	// Subscribe to Lunar reaction events for Gravity accumulation
	c.subscribeToLunarReactions()

	return nil
}

func (c *char) ActionStam(a action.Action, p map[string]int) float64 {
	// Moondew Cleanse does not consume stamina when Verdant Dew >= 1
	if a == action.ActionCharge && c.Core.Player.Verdant.Count() >= 1 {
		return 0
	}
	return c.Character.ActionStam(a, p)
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 10
	}
	return c.Character.AnimationStartDelay(k)
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "gravity":
		return c.gravity, nil
	case "gravity-lc":
		return c.gravityLC, nil
	case "gravity-lb":
		return c.gravityLB, nil
	case "gravity-lcrs":
		return c.gravityLCrs, nil
	case "lunacy":
		return c.lunacyStacks, nil
	case "lunar-domain":
		if c.lunarDomainActive {
			return 1, nil
		}
		return 0, nil
	case "moonridge-dew":
		return c.moonridgeDew, nil
	}
	return c.Character.Condition(fields)
}

// getDominantLunarType returns the Lunar reaction type with the most accumulated Gravity
// Returns: "lc" for Lunar-Charged, "lb" for Lunar-Bloom, "lcrs" for Lunar-Crystallize
func (c *char) getDominantLunarType() string {
	if c.gravityLC >= c.gravityLB && c.gravityLC >= c.gravityLCrs {
		return "lc"
	}
	if c.gravityLB >= c.gravityLC && c.gravityLB >= c.gravityLCrs {
		return "lb"
	}
	return "lcrs"
}
