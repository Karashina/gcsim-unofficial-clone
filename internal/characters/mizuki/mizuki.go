package mizuki

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

func init() {
	core.RegisterCharFunc(keys.Mizuki, NewChar)
}

type char struct {
	*tmpl.Character
	dreamdrifterSrc int
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 99
	c.NormalHitNum = normalHitNum
	c.SkillCon = 99
	c.BurstCon = 99

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.onExitField()
	c.snackHandler("init")
	c.a1()
	c.a4()
	return nil
}

func (c *char) onExitField() {
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(_ ...interface{}) bool {
		if c.StatModIsActive(skillKey) {
			c.DeleteStatMod(skillKey)
		}
		return false
	}, "mizuki-exit")
}
