package varka

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
)

func init() {
	core.RegisterCharFunc(keys.Varka, NewChar)
}

// Status/buff key constants
const (
	sturmUndDrangKey  = "varka-sturm-und-drang"
	fwaCDKey          = "varka-fwa-cd"
	fwaChargeICDKey   = "varka-fwa-charge-icd"
	a1Key             = "varka-a1"
	a4Key             = "varka-a4-azure-fang"
	a4ICDPrefix       = "varka-a4-icd-"
	c1LyricalKey      = "varka-c1-lyrical"
	c4Key             = "varka-c4-buff"
	c6FWAWindowKey    = "varka-c6-fwa-window"
	c6AzureWindowKey  = "varka-c6-azure-window"
	particleICDKey    = "varka-particle-icd"
	normalHitNum      = 5
	sturmNormalHitNum = 5
)

type char struct {
	*tmpl.Character

	// Sturm und Drang state
	sturmActive        bool               // whether in S&D mode
	sturmSrc           int                // source ID to track S&D instance
	otherElement       attributes.Element // determined "Other" element from party
	hasOtherEle        bool               // whether party has Pyro/Hydro/Electro/Cryo
	savedNormalCounter int

	// Four Winds' Ascension
	fwaCharges       int // current FWA charges (max 2, or 3 with C1)
	fwaMaxCharges    int // max charges (2 base, 3 with C1)
	fwaCDEndFrame    int // frame when next FWA charge becomes available
	cdReductionCount int // NA CD reduction counter per S&D activation (max 15)
	cdReductionMax   int // max CD reductions (15 base)

	// A1 party composition
	anemoCount   int     // number of Anemo characters in party
	sameEleCount int     // max count of same element from Pyro/Hydro/Electro/Cryo
	a1MultFactor float64 // 1.0, 1.4, or 2.2

	// A4 stacks
	a4Stacks int
	a4Expiry int

	// Hexerei system
	isHexerei   bool
	hasHexBonus bool // whether 2+ Hexerei characters in party
}

func NewChar(s *core.Core, w *character.CharWrapper, p info.CharacterProfile) error {
	c := char{
		isHexerei:      true,
		fwaMaxCharges:  2,
		cdReductionMax: 15,
	}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.SkillCon = 3
	c.BurstCon = 5

	// Disable hexerei if nohex=1 parameter
	if nohex, ok := p.Params["nohex"]; ok && nohex == 1 {
		c.isHexerei = false
	}

	w.Character = &c
	return nil
}

func (c *char) Init() error {
	// Determine party element composition
	c.determineOtherElement()

	// Determine A1 multiplier
	c.determineA1Mult()

	// Check Hexerei bonus
	c.checkHexereiBonus()

	// Initialize A1 passive (ATK-based DMG bonus)
	c.a1Init()

	// Initialize A4 passive (Swirl subscription)
	if c.Base.Ascension >= 4 {
		c.a4Init()
	}

	// Initialize constellations
	if c.Base.Cons >= 1 {
		c.fwaMaxCharges = 3
	}
	if c.Base.Cons >= 2 {
		c.c2Init()
	}
	if c.Base.Cons >= 4 {
		c.c4Init()
	}

	return nil
}

// ActionReady handles skill availability including FWA in S&D mode
func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	if a == action.ActionSkill && c.sturmActive {
		// In S&D mode, skill becomes Four Winds' Ascension
		c.updateFWACharges()
		if c.fwaCharges > 0 {
			return true, action.NoFailure
		}
		// C6: windows allow skill without FWA charges
		if c.Base.Cons >= 6 && (c.StatusIsActive(c6AzureWindowKey) || c.StatusIsActive(c6FWAWindowKey)) {
			return true, action.NoFailure
		}
		return false, action.SkillCD
	}
	return c.Character.ActionReady(a, p)
}

// updateFWACharges checks if any charges have come off CD
func (c *char) updateFWACharges() {
	for c.fwaCharges < c.fwaMaxCharges && c.Core.F >= c.fwaCDEndFrame {
		c.fwaCharges++
		if c.fwaCharges < c.fwaMaxCharges {
			c.fwaCDEndFrame += 11 * 60
		}
	}
}

// Condition responds to character state queries
func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "hexerei":
		return c.isHexerei, nil
	case "sturm-und-drang":
		return c.sturmActive, nil
	case "fwa-charges":
		return c.fwaCharges, nil
	case "other-element":
		return c.otherElement.String(), nil
	case "a4-stacks":
		if c.Core.F >= c.a4Expiry {
			return 0, nil
		}
		return c.a4Stacks, nil
	}
	return c.Character.Condition(fields)
}

// determineOtherElement finds the highest priority element from party
// Priority: Pyro > Hydro > Electro > Cryo
func (c *char) determineOtherElement() {
	c.hasOtherEle = false
	priorityElements := []attributes.Element{
		attributes.Pyro,
		attributes.Hydro,
		attributes.Electro,
		attributes.Cryo,
	}
	for _, ele := range priorityElements {
		for _, char := range c.Core.Player.Chars() {
			if char.Base.Element == ele {
				c.otherElement = ele
				c.hasOtherEle = true
				return
			}
		}
	}
	// Default to Anemo if no eligible element found
	c.otherElement = attributes.Anemo
}

// determineA1Mult computes the A1 passive multiplier based on party composition
func (c *char) determineA1Mult() {
	if c.Base.Ascension < 1 {
		c.a1MultFactor = 1.0
		return
	}

	c.anemoCount = 0
	eleCounts := map[attributes.Element]int{}
	for _, char := range c.Core.Player.Chars() {
		switch char.Base.Element {
		case attributes.Anemo:
			c.anemoCount++
		case attributes.Pyro, attributes.Hydro, attributes.Electro, attributes.Cryo:
			eleCounts[char.Base.Element]++
		}
	}

	// Find max same-element count among Pyro/Hydro/Electro/Cryo
	c.sameEleCount = 0
	for _, cnt := range eleCounts {
		if cnt > c.sameEleCount {
			c.sameEleCount = cnt
		}
	}

	// Determine multiplier:
	// 2+ Anemo AND 2+ same other element: 2.2x
	// 2+ Anemo OR 2+ same other element: 1.4x
	// Otherwise: 1.0x
	hasAnemoBonus := c.anemoCount >= 2
	hasOtherBonus := c.sameEleCount >= 2

	if hasAnemoBonus && hasOtherBonus {
		c.a1MultFactor = 2.2
	} else if hasAnemoBonus || hasOtherBonus {
		c.a1MultFactor = 1.4
	} else {
		c.a1MultFactor = 1.0
	}
}

// checkHexereiBonus checks if party has 2+ Hexerei characters
func (c *char) checkHexereiBonus() {
	if !c.isHexerei {
		c.hasHexBonus = false
		return
	}
	hexereiCount := 0
	for _, char := range c.Core.Player.Chars() {
		if result, err := char.Condition([]string{"hexerei"}); err == nil {
			if isHex, ok := result.(bool); ok && isHex {
				hexereiCount++
			}
		}
	}
	c.hasHexBonus = hexereiCount >= 2
}

// getCDReductionAmount returns the CD reduction per NA hit based on Hexerei bonus
func (c *char) getCDReductionAmount() int {
	if c.hasHexBonus {
		return 60 // 1.0s with Hexerei Secret Rite
	}
	return 30 // 0.5s base
}
