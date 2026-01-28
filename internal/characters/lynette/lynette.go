package lynette

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
)

func init() {
	core.RegisterCharFunc(keys.Lynette, NewChar)
}

type char struct {
	*tmpl.Character
	wp             *ReactableWeapon
	a1Buff         []float64
	skillAI        combat.AttackInfo
	skillAlignedAI combat.AttackInfo
	shadowsignSrc  int
	vividCount     int
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 70
	c.NormalHitNum = normalHitNum
	c.BurstCon = 3
	c.SkillCon = 5
	c.HasArkhe = true

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.skillAI = combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Enigmatic Feint",
		AttackTag:          attacks.AttackTagElementalArt,
		ICDTag:             attacks.ICDTagNone,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeSpear,
		Element:            attributes.Anemo,
		Durability:         25,
		Mult:               skill[c.TalentLvlSkill()],
		HitlagFactor:       0.01,
		CanBeDefenseHalted: true,
	}
	c.skillAlignedAI = combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Surging Blade (" + c.Base.Key.Pretty() + ")",
		AttackTag:          attacks.AttackTagElementalArt,
		ICDTag:             attacks.ICDTagNone,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeSpear,
		Element:            attributes.Anemo,
		Durability:         0,
		Mult:               skillAligned[c.TalentLvlSkill()],
		HitlagFactor:       0.01,
		CanBeDefenseHalted: true,
	}

	c.a1Setup()

	c.vividCount = 1
	c.c2()
	c.c4()
	c.WeaponReactionHandler()
	c.onExitField()

	return nil
}

func (c *char) onExitField() {
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(_ ...interface{}) bool {
		if c.StatModIsActive(weaponoutkey) {
			c.DeleteStatus(weaponoutkey)
		}
		return false
	}, "lynette-exit")
}

func (c *char) WeaponReactionHandler() {
	c.Core.Events.Subscribe(event.OnInitialize, func(args ...interface{}) bool {
		c.wp = c.newReactableWeapons()
		return false
	}, "lynette-weaponreactionhandler-init")

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
		if ele == attributes.Hydro && !c.StatusIsActive(weaponoutkey) && !c.StatusIsActive("bennett-field") {
			return false // for whatever reason, candace infusion will not work if she is not holding weapon
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
	}, "lynette-infusion")
}
