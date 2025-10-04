package flins

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Flins, NewChar)
}

type char struct {
	*tmpl.Character
	northlandCD int
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)
	c.SkillCon = 3
	c.BurstCon = 5

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.AddStatus("moonsignKey", -1, false)
	c.moonsignInitFunc()
	c.InitLCallback()
	c.a0()
	c.a1()
	c.a4()
	return nil
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 11
	}
	return c.Character.AnimationStartDelay(k)
}

func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	if a == action.ActionSkill && c.StatusIsActive(skillKey) {
		if c.StatusIsActive(northlandCdKey) {
			return false, action.SkillCD // Fails if Northland Spearstorm is still on CD (This CD is unaffected by other effects such as c.CDReduction())
		}
		return true, action.NoFailure // Make Northland Spearstorm usable even on normal skill is in CD
	}
	if a == action.ActionBurst && c.StatusIsActive(northlandKey) {
		if !c.Core.Flags.IgnoreBurstEnergy && c.Energy < 30 {
			return false, action.InsufficientEnergy // Energy cost of Thunderous Symphony is 30
		}
		return true, action.NoFailure // Make Thunderous Symphony usable even on normal burst is in CD
	}
	return c.Character.ActionReady(a, p)
}

func (c *char) moonsignInitFunc() {
	count := 0
	for _, char := range c.Core.Player.Chars() {
		if char.StatusIsActive("moonsignKey") {
			count++
		}
	}
	switch count {
	case 1:
		c.MoonsignNascent = true // Moonsign: Nascent Gleam
		c.MoonsignAscendant = false
	case 2, 3, 4:
		c.MoonsignAscendant = true // Moonsign: Ascendant Gleam
		c.MoonsignNascent = false
	default:
		c.MoonsignNascent = false
		c.MoonsignAscendant = false
	}
}
