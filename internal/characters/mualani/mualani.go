package mualani

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/core/reactions"
	"github.com/genshinsim/gcsim/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Mualani, NewChar)
}

type char struct {
	*tmpl.Character
	WaveMomentum int
	pufferCount  int
	a4stacks     int
	a4buff       float64
	c1count      int
	c1buff       float64
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.CharZone = info.ZoneNatlan
	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.SkillCon = 3
	c.BurstCon = 5

	c.NightsoulPointMax = 60
	c.NightsoulPoint = 0
	c.HasNightsoul = true
	c.OnNightsoul = false

	c.a4stacks = 0
	c.a4buff = 0
	c.c1buff = 0

	c.SetNumCharges(action.ActionSkill, 1)

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.NightsoulBurst()
	c.a4()
	c.c4()
	return nil
}

func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	if a == action.ActionSkill && c.StatusIsActive(skillKey) {
		return true, action.NoFailure
	}
	return c.Character.ActionReady(a, p)
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "nightsoul-points":
		if c.NightsoulPoint <= 0 {
			return 0, nil
		}
		return c.NightsoulPoint, nil
	default:
		return c.Character.Condition(fields)
	}
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	switch k {
	case model.AnimationXingqiuN0StartDelay:
		if c.StatusIsActive(skillKey) {
			return 12
		}
		return 0
	default:
		return c.Character.AnimationStartDelay(k)
	}
}

func (c *char) NightsoulBurst() {
	natlancount := 0
	for _, this := range c.Core.Player.Chars() {
		if this.CharZone == info.ZoneNatlan {
			natlancount++
		}
	}
	cd := -1
	switch natlancount {
	case 1:
		cd = 18 * 60
	case 2:
		cd = 12 * 60
	case 3:
		cd = 9 * 60
	case 4:
		cd = 9 * 60
	default:
		cd = -1
	}

	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if c.StatusIsActive("NightsoulBurst-ICD") {
			return false
		}
		if atk.Info.Element == attributes.Physical {
			return false
		}
		if atk.Info.Abil == string(reactions.Burning) || atk.Info.Abil == string(reactions.ElectroCharged) {
			return false
		}

		for _, p := range c.Core.Player.Chars() {
			p.AddStatus("NightsoulBurst-ICD", cd, true)
		}
		c.Core.Events.Emit(event.OnNightsoulBurst)

		return false
	}, "Mualani-NightsoulBurst")

}
