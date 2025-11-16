package kazuha

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
	core.RegisterCharFunc(keys.Kazuha, NewChar)
}

type char struct {
	*tmpl.Character
	wp                    *ReactableWeapon
	a1Absorb              attributes.Element
	a1AbsorbCheckLocation combat.AttackPattern
	qAbsorb               attributes.Element
	qFieldSrc             int
	qAbsorbCheckLocation  combat.AttackPattern
	qTickSnap             combat.Snapshot
	qTickAbsorbSnap       combat.Snapshot
	c2buff                []float64
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 60
	c.BurstCon = 5
	c.SkillCon = 3
	c.NormalHitNum = normalHitNum

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.a1Absorb = attributes.NoElement
	c.a4()

	// make sure to use the same key everywhere so that these passives don't stack
	c.Core.Player.AddStamPercentMod("utility-dash", -1, func(a action.Action) (float64, bool) {
		if a == action.ActionDash && c.CurrentHPRatio() > 0 {
			return -0.2, false
		}
		return 0, false
	})

	if c.Base.Cons >= 2 {
		c.c2buff = make([]float64, attributes.EndStatType)
		c.c2buff[attributes.EM] = 200
	}
	c.WeaponReactionHandler()
	return nil
}

func (c *char) Condition(fields []string) (any, error) {
	t := c.Core.Combat.PrimaryTarget().(core.Reactable)
	switch fields[0] {
	case "kazuhaeye-pyro":
		if t.AuraContains(attributes.Pyro) {
			return 1, nil
		}
		return 0, nil
	case "kazuhaeye-hydro":
		if t.AuraContains(attributes.Hydro) {
			return 1, nil
		}
		return 0, nil
	case "kazuhaeye-electro":
		if t.AuraContains(attributes.Electro) {
			return 1, nil
		}
		return 0, nil
	case "kazuhaeye-cryo":
		if t.AuraContains(attributes.Cryo) {
			return 1, nil
		}
		return 0, nil
	default:
		return c.Character.Condition(fields)
	}
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 10
	}
	return c.Character.AnimationStartDelay(k)
}

func (c *char) WeaponReactionHandler() {
	c.Core.Events.Subscribe(event.OnInitialize, func(args ...interface{}) bool {
		c.wp = c.newReactableWeapons()
		return false
	}, "kazuha-weaponreactionhandler-init")

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
	}, "kazuha-infusion")
}

