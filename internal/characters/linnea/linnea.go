package linnea

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Linnea, NewChar)
}

// Lumi form type
type lumiForm int

const (
	lumiFormNone     lumiForm = iota
	lumiFormSuper             // Super Power Form
	lumiFormUltimate          // Ultimate Power Form
	lumiFormStandard          // Standard Power Form
)

type char struct {
	*tmpl.Character
	// Lumi state
	lumiActive   bool
	lumiSrc      int
	lumiForm     lumiForm
	lumiTickSrc  int
	lumiComboIdx int // combo position for Super Power Form (0,1=punch, 2=hammer)
	// constellation state
	fieldCatalogStacks int
	fieldCatalogSrc    int
	c2MoondriftSrc     int
}

const (
	lumiKey              = "linnea-lumi-active"
	burstHealKey         = "linnea-burst-heal"
	particleICDKey       = "linnea-particle-icd"
	fieldCatalogKey      = "linnea-field-catalog"
	c2CritDmgKey         = "linnea-c2-critdmg"
	c4DefKey             = "linnea-c4-def"
	a1GeoResKey          = "linnea-a1-geo-res"
	a1GeoResAscendKey    = "linnea-a1-geo-res-ascend"
	c4DefActiveKey       = "linnea-c4-def-active"
	lumiDuration         = 1560 // tE->end: 1560f (26.0s from skill start)
	lumiSuperTickRate    = 141  // Super Form: PPP 2hit -> next PPP 1hit = 120f + inter-hit 21f
	lumiSuperPPPToHOH    = 109  // PPP->HOH: inter-PPP hit (21f) + PPP hit2->HOH hit (88f) = 109f
	lumiSuperHOHToPPP    = 61   // HOH->PPP: HOH hit -> next PPP hit1 = 61f
	lumiStandardTickRate = 321  // Standard Form: PPP 2hit -> next PPP 1hit = 300f + inter-hit 21f
	// first tick delay (tap E -> Lumi 1st hit: 108f, Q -> Lumi 1st hit: 106f)
	lumiFirstTickFromE        = 108                                                   // tap E -> Lumi first hit
	lumiFirstTickFromQ        = 106                                                   // Q -> Lumi first hit (after Burst activation)
	lumiStdFirstTickAfterMash = 243                                                   // mE->hitmark(111) + hitmark->PPP1(132)
	skillCD                   = 18 * 60                                               // 18s
	burstCD                   = 15 * 60                                               // 15s
	burstCDDelay              = 2                                                     // CD start: 2f
	burstInitHealDelay        = 96                                                    // Q -> initial heal: 96f
	burstContHealStart        = 158                                                   // Q -> continuous heal start: 158f
	burstHealTickRate         = 60                                                    // heal interval: 60f (1s)
	burstHealTicks            = 12                                                    // number of heal ticks: 12
	burstHealDuration         = burstContHealStart + burstHealTicks*burstHealTickRate // heal status duration
	fieldCatalogDuration      = 10 * 60                                               // Field Catalog duration: 10s
	maxFieldCatalog           = 18
)

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.BurstCon = 5
	c.SkillCon = 3

	// initialize state
	c.lumiActive = false
	c.lumiSrc = -1
	c.lumiForm = lumiFormNone
	c.lumiTickSrc = -1
	c.lumiComboIdx = 0
	c.fieldCatalogStacks = 0

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	// initialize A0 passive (LCrs key, moonsign level, DEF-based LCrs bonus)
	c.a0Init()
	// initialize A1 passive (Geo RES reduction)
	c.a1Init()
	// initialize A4 passive (Elemental Mastery buff)
	if c.Base.Ascension >= 4 {
		c.a4Init()
	}
	// initialize constellations
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
	return nil
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	return c.Character.AnimationStartDelay(k)
}

// Condition queries character state from GCSL
func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "lumi-active":
		if c.lumiActive {
			return 1, nil
		}
		return 0, nil
	case "lumi-form":
		return int(c.lumiForm), nil
	case "field-catalog":
		return c.fieldCatalogStacks, nil
	case "moonsign-ascendant":
		if c.MoonsignAscendant {
			return 1, nil
		}
		return 0, nil
	}
	return c.Character.Condition(fields)
}

// onMoondriftHarmony handles Moondrift Harmony activation (triggers C1/C2/C4 effects)
func (c *char) onMoondriftHarmony() {
	// C1: add Field Catalog stacks
	if c.Base.Cons >= 1 {
		c.c1OnHarmony()
	}
	// C2: CRIT DMG bonus for Hydro/Geo party members
	if c.Base.Cons >= 2 {
		c.c2OnHarmony()
	}
	// C4: DEF increase
	if c.Base.Cons >= 4 {
		c.c4OnHarmony()
	}
}
