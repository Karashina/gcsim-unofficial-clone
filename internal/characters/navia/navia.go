package navia

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Navia, NewChar)
}

type char struct {
	*tmpl.Character
	shrapnel int
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)
	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.SkillCon = 3
	c.BurstCon = 5
	c.HasArkhe = true

	c.SetNumCharges(action.ActionSkill, 2)
	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.a4()
	c.shrapnelGain()
	return nil
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "shrapnel":
		return c.shrapnel, nil
	default:
		return c.Character.Condition(fields)
	}
}
func (c *char) Snapshot(ai *combat.AttackInfo) combat.Snapshot {
	ds := c.Character.Snapshot(ai)

	if c.Character.StatusIsActive(a1Key) { // weapon infusion can't be overriden for navia
		switch ai.AttackTag {
		case attacks.AttackTagNormal:
		case attacks.AttackTagPlunge:
		case attacks.AttackTagExtra:
		default:
			return ds
		}
		ai.Element = attributes.Geo
	}
	return ds
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	switch k {
	case model.AnimationXingqiuN0StartDelay:
		return 19
	case model.AnimationYelanN0StartDelay:
		return 19
	default:
		return c.Character.AnimationStartDelay(k)
	}
}

