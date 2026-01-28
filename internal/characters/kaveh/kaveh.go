package kaveh

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Kaveh, NewChar)
}

type char struct {
	*tmpl.Character
	a4Stacks int
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.BurstCon = 3
	c.SkillCon = 5
	c.EnergyMax = 80
	c.NormalHitNum = len(attackHitmarks)

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	if c.Base.Cons >= 4 {
		c.c4()
	}
	if c.Base.Cons >= 6 {
		c.c6()
	}
	c.a1()
	c.a4AddStacksHandler()
	c.addBurstExitHandler()
	return nil
}

func (c *char) Snapshot(ai *combat.AttackInfo) combat.Snapshot {
	ds := c.Character.Snapshot(ai)

	if c.StatModIsActive(burstKey) {
		switch ai.AttackTag {
		case attacks.AttackTagNormal,
			attacks.AttackTagPlunge,
			attacks.AttackTagExtra:
			ai.Element = attributes.Dendro
		}
	}

	return ds
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	switch k {
	case model.AnimationXingqiuN0StartDelay:
		return 25
	case model.AnimationYelanN0StartDelay:
		return 23
	default:
		return c.Character.AnimationStartDelay(k)
	}
}
