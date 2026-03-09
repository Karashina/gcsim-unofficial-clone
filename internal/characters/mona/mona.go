package mona

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
)

const (
	bubbleKey = "mona-bubble"
	omenKey   = "omen-debuff"
)

func init() {
	core.RegisterCharFunc(keys.Mona, NewChar)
}

type char struct {
	*tmpl.Character
	a4Stats  []float64
	c2icd    int
	c6Src    int
	c6Stacks int
	// Hexereiモード（nohex=1が指定されない限りデフォルトtrue）
	isHexerei bool
}

func NewChar(s *core.Core, w *character.CharWrapper, p info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.BurstCon = 3
	c.SkillCon = 5

	c.c2icd = -1

	// nohex=1が指定されない限りデフォルトはHexereiキャラクター
	c.isHexerei = true
	if nohex, ok := p.Params["nohex"]; ok && nohex == 1 {
		c.isHexerei = false
	}

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.burstHook()
	c.burstDamageBonus()
	c.a4()
	if c.Base.Cons >= 1 {
		c.c1()
	}
	if c.Base.Cons >= 4 {
		c.c4()
	}
	return nil
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "hexerei":
		return c.isHexerei, nil
	default:
		return c.Character.Condition(fields)
	}
}
