package illuga

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Illuga, NewChar)
}

type char struct {
	*tmpl.Character
	// Burst state
	orioleSongActive        bool
	orioleSongSrc           int
	nightingaleSongStacks   int
	c2StackCounter          int // Track stacks consumed for C2
	geoConstructBonusStacks int // Track additional stacks from Geo Constructs (max 15)
	// Moonsign state
	moonsignAscendant bool
	// A4 party composition tracking
	a4HydroCount int
	a4GeoCount   int
}

const (
	orioleSongKey      = "illuga-oriole-song-active"
	lightkeeperOathKey = "illuga-lightkeeper-oath"
	particleICDKey     = "illuga-particle-icd"
)

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.BurstCon = 5
	c.SkillCon = 3

	// Initialize state
	c.orioleSongActive = false
	c.orioleSongSrc = -1
	c.nightingaleSongStacks = 0
	c.c2StackCounter = 0
	c.moonsignAscendant = false

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	// Initialize A4 passive (party composition tracking)
	c.a4Init()

	// Initialize constellations
	c.consInit()

	// Grant moonsignKey for A0 (Moonsign Level +1)
	c.AddStatus("moonsignKey", -1, true)

	// Subscribe to moonsign state updates
	c.updateMoonsignState()

	return nil
}

// updateMoonsignState updates Illuga's internal moonsign state based on party-wide flags
func (c *char) updateMoonsignState() {
	c.moonsignAscendant = c.MoonsignAscendant
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 8
	}
	return c.Character.AnimationStartDelay(k)
}

// Condition allows querying character state from GCSL
func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "oriole-song":
		if c.orioleSongActive {
			return 1, nil
		}
		return 0, nil
	case "nightingale-stacks":
		return c.nightingaleSongStacks, nil
	case "moonsign-ascendant":
		if c.moonsignAscendant {
			return 1, nil
		}
		return 0, nil
	}
	return c.Character.Condition(fields)
}

// isMoonsignAscendant checks if Moonsign is Ascendant Gleam
// Queries the party-wide Moonsign status set during initialization
func (c *char) isMoonsignAscendant() bool {
	return c.MoonsignAscendant
}
