package sigewinne

import (
	"math"

	"github.com/genshinsim/gcsim/internal/common"
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/core/targets"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

var skillFrames []int

const (
	skillKey          = "sigewinne-skill"
	skillAbilKey      = "Bolstering Bubblebalm DMG"
	arkheICDKey       = "sigewinne-arkhe-icd"
	skillInterval     = 113
	skillarkheHitmark = 27
)

func init() {
	skillFrames = frames.InitAbilSlice(33)
	skillFrames[action.ActionAttack] = 33
	skillFrames[action.ActionAim] = 33
	skillFrames[action.ActionBurst] = 33
	skillFrames[action.ActionDash] = 33
	skillFrames[action.ActionJump] = 33
	skillFrames[action.ActionSwap] = 33
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	switch p["hold"] {
	case 1:
		c.skillsize = 1
		c.c2skillsize = 3
		c.skillHitmark = 75
	case 2:
		c.skillsize = 2
		c.c2skillsize = 3
		c.skillHitmark = 109
	default:
		c.skillsize = 0
		c.skillHitmark = 40
	}

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       skillAbilKey,
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagElementalArt,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
		FlatDmg:    skill[c.TalentLvlSkill()] * c.MaxHP(),
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 1.5),
		c.skillHitmark,
		c.skillHitmark,
		c.particleCB,
		c.c2Resist,
		c.arkhe(c.skillHitmark+skillarkheHitmark),
	)
	c.makedrop()
	c.QueueCharTask(c.skillheal, c.skillHitmark)
	c.a1()
	var count int
	if c.Base.Cons < 1 {
		count = 5
	} else {
		count = 8
	}

	for i := 1; i < count; i++ {
		c.QueueCharTask(c.skillProc, c.skillHitmark+i*skillInterval)
		c.QueueCharTask(c.skillheal, c.skillHitmark+i*skillInterval)
		if i == count-1 {
			c.QueueCharTask(c.selfheal, i*skillInterval)
		}
	}

	c.SetCDWithDelay(action.ActionSkill, 18*60, 0)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionAim], // earliest cancel
		State:           action.SkillState,
		OnRemoved:       func(next action.AnimationState) { c.c2Remove() },
	}, nil
}

func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	count := 4.0
	c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Hydro, c.ParticleDelay)
}

func (c *char) skillProc() {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       skillAbilKey,
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
		FlatDmg:    skill[c.TalentLvlSkill()] * c.MaxHP(),
	}

	var pos geometry.Point
	area := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 10)
	enemy := c.Core.Combat.RandomEnemyWithinArea(
		area,
		nil,
	)
	if enemy != nil {
		pos = enemy.Pos()
	} else {
		pos = c.Core.Combat.PrimaryTarget().Pos()
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(pos, nil, 1.5),
		1,
		1,
		c.c2Resist,
	)
}

func (c *char) skillheal() {
	// heal other chars
	for _, char := range c.Core.Player.Chars() {
		if char.Index != c.Index {
			c.Core.Player.Heal(info.HealInfo{
				Caller:  c.Index,
				Target:  char.Index,
				Message: "Bolstering Bubblebalm Healing",
				Src:     c.MaxHP()*skillheal[c.TalentLvlSkill()] + skillbonus[c.TalentLvlSkill()],
				Bonus:   c.Stat(attributes.Heal) + float64(c.skillsize)*0.05,
			})
		}
	}
	c.skillsizemanager()
	if c.Base.Cons >= 1 {
		c.Tags[a1BuffKey]++
	}
}

func (c *char) selfheal() {
	c.Core.Player.Heal(info.HealInfo{
		Caller:  c.Index,
		Target:  c.Index,
		Message: "Final Bounce Healing",
		Src:     c.MaxHP() * 0.5,
		Bonus:   c.Stat(attributes.Heal),
	})
}

func (c *char) arkhe(delay int) combat.AttackCBFunc {
	// triggers on hitting anything, not just enemy
	return func(a combat.AttackCB) {
		if c.StatusIsActive(arkheICDKey) {
			return
		}
		c.AddStatus(arkheICDKey, 10*60, true)

		aiArkhe := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Surging Blade (" + c.Base.Key.Pretty() + ")",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Hydro,
			Durability: 0,
			FlatDmg:    skillaligned[c.TalentLvlSkill()] * c.MaxHP(),
		}

		c.Core.QueueAttack(
			aiArkhe,
			combat.NewCircleHitOnTarget(a.Target.Pos(), nil, 2),
			delay,
			delay,
		)
	}
}

func (c *char) skillbuff() {

	m := make([]float64, attributes.EndStatType)
	m[attributes.Hydro] = 0.05 * float64(c.skillsize)

	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("sigewinne-skill-buff", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagElementalArt {
				return nil, false
			}
			return m, true
		},
	})
}

func (c *char) skillsizemanager() {
	if c.Base.Cons >= 2 && c.c2skillsize > 0 {
		c.c2skillsize--
	} else {
		c.skillsize--
		if c.skillsize < 0 {
			c.skillsize = 0
		}
	}
}

func (c *char) makedrop() {
	droplet1 := c.newDroplet()
	droplet2 := c.newDroplet()
	c.Core.Combat.AddGadget(droplet1)
	c.Core.Combat.AddGadget(droplet2)
}

func (c *char) dropletPickUp(count int) {
	for _, g := range c.Core.Combat.Gadgets() {
		if count == 0 {
			return
		}

		droplet, ok := g.(*common.SourcewaterDroplet)
		if !ok {
			continue
		}
		droplet.Kill()
		count--

		c.ModifyHPDebtByAmount(0.1 * c.MaxHP())
	}
}

func (c *char) newDroplet() *common.SourcewaterDroplet {
	player := c.Core.Combat.Player()
	pos := geometry.CalcRandomPointFromCenter(
		geometry.CalcOffsetPoint(
			player.Pos(),
			geometry.Point{Y: 3.5},
			player.Direction(),
		),
		0.3,
		3,
		c.Core.Rand,
	)
	droplet := common.NewSourcewaterDroplet(c.Core, pos, combat.GadgetTypSourcewaterDropletHydroTrav)
	return droplet
}

func (c *char) bolmanager() {
	c.Core.Events.Subscribe(event.OnHPDebt, func(args ...interface{}) bool {
		index := args[0].(int)
		amount := args[1].(float64)
		if index != c.Index {
			return false
		}
		if amount >= 0 {
			return false
		}

		energyamt := min(5, math.Abs(amount)/2000)

		c.AddEnergy("sigewinne-bol", energyamt)

		return false
	}, "sigewinne-bol")
}
