package diluc

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Diluc, NewChar)
}

type char struct {
	*tmpl.Character
	wp                 *ReactableWeapon
	eCounter           int
	a4buff             []float64
	c2buff             []float64
	c2stack            int
	c4buff             []float64
	savedNormalCounter int
	c6Count            int
}

const eWindowKey = "diluc-e-window"

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 40
	c.NormalHitNum = normalHitNum
	c.SkillCon = 3
	c.BurstCon = 5

	c.eCounter = 0

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.a4buff = make([]float64, attributes.EndStatType)
	c.a4buff[attributes.PyroP] = 0.2

	if c.Base.Cons >= 1 && c.Core.Combat.DamageMode {
		c.c1()
	}
	if c.Base.Cons >= 2 {
		c.c2()
	}
	if c.Base.Cons >= 4 {
		c.c4buff = make([]float64, attributes.EndStatType)
		c.c4buff[attributes.DmgP] = 0.4
	}
	c.WeaponReactionHandler()
	return nil
}

func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	// check if it is possible to use next skill
	if a == action.ActionSkill && c.StatusIsActive(eWindowKey) {
		return true, action.NoFailure
	}
	return c.Character.ActionReady(a, p)
}

// pyro infuse can't be overwritter
func (c *char) Snapshot(ai *combat.AttackInfo) combat.Snapshot {
	ds := c.Character.Snapshot(ai)

	if c.StatusIsActive(burstBuffKey) {
		// infusion to attacks only
		switch ai.AttackTag {
		case attacks.AttackTagNormal:
		case attacks.AttackTagPlunge:
		case attacks.AttackTagExtra:
		default:
			return ds
		}
		ai.Element = attributes.Pyro
	}

	return ds
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 15
	}
	return c.Character.AnimationStartDelay(k)
}

func (c *char) WeaponReactionHandler() {
	c.Core.Events.Subscribe(event.OnInitialize, func(args ...interface{}) bool {
		c.wp = c.newReactableWeapons()
		return false
	}, "diluc-weaponreactionhandler-init")

	c.Core.Events.Subscribe(event.OnInfusion, func(args ...interface{}) bool {
		index := args[0].(int)
		ele := args[1].(attributes.Element)
		dur := args[2].(int)
		if c.Core.Player.ActiveChar().Index != c.Index {
			return false
		}
		if index != c.Index {
			return false
		}
		infai := combat.AttackInfo{
			ActorIndex: index,
			Abil:       "Weapon Infusion",
			Element:    ele,
			Durability: 25,
		}
		infae := combat.AttackEvent{
			Info:        infai,
			Pattern:     combat.NewSingleTargetHit(0),
			SourceFrame: c.Core.F,
		}
		c.wp.weaponreact(&infae)
		c.QueueCharTask(c.wp.resetgauge, dur)
		return false
	}, "diluc-infusion")
}
