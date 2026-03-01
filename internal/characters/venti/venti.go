package venti

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Venti, NewChar)
}

type char struct {
	*tmpl.Character
	qPos                geometry.Point
	qAbsorb             attributes.Element
	absorbCheckLocation combat.AttackPattern
	aiAbsorb            combat.AttackInfo
	snapAbsorb          combat.Snapshot
	// Hexerei mode (default true unless nohex=1)
	isHexerei   bool
	hasHexBonus bool // 2+ hexerei characters in party

	// Hexerei burst eye tracking
	burstEnd       int // absolute frame when burst eye expires
	normalHexCount int // number of hex normal attack triggers per burst (max 2)
	lastHexTrigger int // last frame hex normal attack trigger fired

	// C1 hexerei: Stormwind Arrow split tracking (0.25s ICD = 15 frames)
	lastStormwindSplit int
}

func NewChar(s *core.Core, w *character.CharWrapper, p info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.BurstCon = 3
	c.SkillCon = 5

	// Default is Hexerei character unless nohex=1 is specified
	c.isHexerei = true
	if nohex, ok := p.Params["nohex"]; ok && nohex == 1 {
		c.isHexerei = false
	}

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	// Check Hexerei bonus (2+ hexerei characters in party)
	c.checkHexereiBonus()

	// A0: Hexerei Secret Rite passive (swirl → DMG buff + burst boost)
	c.a0HexereiInit()

	// C4 (original): Venti gains Anemo DMG +25% for 10s when picking up particle
	if c.Base.Cons >= 4 {
		c.c4Old()
	}

	// C6: persistent AttackMod for CRIT DMG bonus against burst-affected enemies (hexerei only)
	if c.Base.Cons >= 6 {
		c.c6AttackModInit()
	}
	return nil
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 9
	}
	return c.Character.AnimationStartDelay(k)
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "hexerei":
		return c.isHexerei, nil
	default:
		return c.Character.Condition(fields)
	}
}

// checkHexereiBonus determines if party has 2+ hexerei characters.
func (c *char) checkHexereiBonus() {
	if !c.isHexerei {
		c.hasHexBonus = false
		return
	}
	hexereiCount := 0
	for _, ch := range c.Core.Player.Chars() {
		if result, err := ch.Condition([]string{"hexerei"}); err == nil {
			if isHex, ok := result.(bool); ok && isHex {
				hexereiCount++
			}
		}
	}
	c.hasHexBonus = hexereiCount >= 2
}
