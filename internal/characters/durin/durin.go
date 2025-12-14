package durin

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
)

func init() {
	core.RegisterCharFunc(keys.Durin, NewChar)
}

// char is the implementation of Durin character
// A Pyro element character with transmutation states and dragon summoning abilities
type char struct {
	*tmpl.Character

	// State management
	stateDenial bool // true = Denial of Darkness, false = Confirmation of Purity

	// Dragon tracking
	dragonWhiteFlame bool // Whether Dragon of White Flame is summoned
	dragonDarkDecay  bool // Whether Dragon of Dark Decay is summoned
	dragonExpiry     int  // Expiration frame for active dragon
	dragonSrc        int  // Source ID to track dragon instance (prevents old dragons from attacking)

	// A4 Talent: Chaos Formed Like the Night (Primordial Fusion)
	primordialFusionStacks int // Primordial Fusion stack count (max 10)
	primordialFusionExpiry int // Stack expiration frame

	// C1 Constellation: Adamah's Redemption (Cycle of Enlightenment)
	cycleStacks map[int]int // Character index -> Cycle of Enlightenment stack count
	cycleExpiry map[int]int // Character index -> Expiration frame

	// Hexerei system (special character class)
	isHexerei bool // Whether this character has Hexerei attribute

	// Skill internal cooldowns (ICD)
	lastEnergyRestoreFrame int  // Last frame when energy restore was triggered
	particleIcd            bool // Particle generation ICD status
}

// NewChar creates a new instance of Durin character
func NewChar(s *core.Core, w *character.CharWrapper, p info.CharacterProfile) error {
	c := char{
		cycleStacks:            make(map[int]int),
		cycleExpiry:            make(map[int]int),
		lastEnergyRestoreFrame: -9999, // Initialize to far past frame
		isHexerei:              true,  // Default to having Hexerei attribute
	}
	c.Character = tmpl.NewWithWrapper(s, w)

	// Character basic parameters
	c.EnergyMax = 70   // Energy required for Elemental Burst
	c.NormalHitNum = 4 // Number of normal attack hits
	c.SkillCon = 5     // Constellation that increases Elemental Skill talent level
	c.BurstCon = 3     // Constellation that increases Elemental Burst talent level

	// Disable Hexerei attribute if nohex=1 parameter is specified
	if nohex, ok := p.Params["nohex"]; ok && nohex == 1 {
		c.isHexerei = false
	}

	w.Character = &c
	return nil
}

// Init initializes the character and sets up talent and constellation effects
func (c *char) Init() error {
	// Initialize talent effects
	c.a1() // A1: Light Manifest of the Divine Calculus

	// Initialize constellation effects
	if c.Base.Cons >= 1 {
		c.c1() // C1: Adamah's Redemption
	}
	if c.Base.Cons >= 2 {
		c.c2() // C2: Unsound Visions
	}
	if c.Base.Cons >= 4 {
		c.c4() // C4: Emanare's Source
	}

	return nil
}

// ActionReady checks if an action is available
// Handles double skill mechanics: Essential Transmutation can be recasted immediately
func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	// Allow recasting skill immediately if in Essential Transmutation state and recast CD not active
	if a == action.ActionSkill && c.StatusIsActive(essentialTransmutationKey) {
		if c.StatusIsActive(skillRecastCDKey) {
			// Already used recast, must wait for main skill CD
			return false, action.SkillCD
		}
		// Recast allowed even if normal skill CD is active
		return true, action.NoFailure
	}

	return c.Character.ActionReady(a, p)
}

// Condition responds to character state queries
// Allows checking transmutation state, dragon state, stack counts, etc.
func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "state": // Returns current transmutation state as string
		if c.stateDenial {
			return "denial", nil // Denial of Darkness
		}
		return "confirmation", nil // Confirmation of Purity
	case "denial": // Whether in Denial of Darkness state
		return c.stateDenial, nil
	case "confirmation": // Whether in Confirmation of Purity state
		return !c.stateDenial, nil
	case "dragon-white-flame": // Whether Dragon of White Flame is active
		return c.dragonWhiteFlame && c.Core.F < c.dragonExpiry, nil
	case "dragon-dark-decay": // Whether Dragon of Dark Decay is active
		return c.dragonDarkDecay && c.Core.F < c.dragonExpiry, nil
	case "primordial-fusion-stacks": // A4 Primordial Fusion stack count
		if c.Core.F >= c.primordialFusionExpiry {
			return 0, nil
		}
		return c.primordialFusionStacks, nil
	case "hexerei": // Whether has Hexerei attribute
		return c.isHexerei, nil
	case "cycle-stacks": // C1 Cycle of Enlightenment stack count (current active character)
		charIndex := c.Core.Player.ActiveChar().Index
		if expiry, ok := c.cycleExpiry[charIndex]; ok && c.Core.F < expiry {
			return c.cycleStacks[charIndex], nil
		}
		return 0, nil
	}
	return c.Character.Condition(fields)
}

// hasHexereiBonus checks if party has 2 or more Hexerei characters
// Hexerei Bonus: When 2+ Hexerei characters are in party, various effects are increased by 75% (multiplied by 1.75)
// Eligible characters: Durin, Albedo, Klee, Venti, Mona, Fischl, Razor, Sucrose
func (c *char) hasHexereiBonus() bool {
	// No bonus if this character doesn't have Hexerei attribute
	if !c.isHexerei {
		return false
	}

	// Count party members with Hexerei attribute
	hexereiCount := 0
	for _, char := range c.Core.Player.Chars() {
		if result, err := char.Condition([]string{"hexerei"}); err == nil {
			if isHexerei, ok := result.(bool); ok && isHexerei {
				hexereiCount++
			}
		}
	}

	// Bonus activates with 2 or more Hexerei characters
	return hexereiCount >= 2
}
