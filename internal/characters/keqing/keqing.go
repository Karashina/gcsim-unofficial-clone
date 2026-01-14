package keqing

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Keqing, NewChar)
}

type char struct {
	*tmpl.Character
	wp     *ReactableWeapon
	a4buff []float64
	c4buff []float64
	c6buff []float64
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 40
	c.NormalHitNum = normalHitNum
	c.BurstCon = 3
	c.SkillCon = 5

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.a4buff = make([]float64, attributes.EndStatType)
	c.a4buff[attributes.CR] = 0.15
	c.a4buff[attributes.ER] = 0.15

	if c.Base.Cons >= 4 {
		c.c4buff = make([]float64, attributes.EndStatType)
		c.c4buff[attributes.ATKP] = 0.25
		c.c4()
	}
	if c.Base.Cons >= 6 {
		c.c6buff = make([]float64, attributes.EndStatType)
		c.c6buff[attributes.ElectroP] = 0.06
	}
	c.WeaponReactionHandler()
	return nil
}

func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	// check if stiletto is on-field
	if a == action.ActionSkill && c.Core.Status.Duration(stilettoKey) > 0 {
		return true, action.NoFailure
	}
	return c.Character.ActionReady(a, p)
}

func (c *char) ActionStam(a action.Action, p map[string]int) float64 {
	if a == action.ActionCharge {
		return 25
	}
	return c.Character.ActionStam(a, p)
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 8
	}
	return c.Character.AnimationStartDelay(k)
}

func (c *char) WeaponReactionHandler() {
	c.Core.Events.Subscribe(event.OnInitialize, func(args ...interface{}) bool {
		c.wp = c.newReactableWeapons()
		return false
	}, "keqing-weaponreactionhandler-init")

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
	}, "keqing-infusion")
}

