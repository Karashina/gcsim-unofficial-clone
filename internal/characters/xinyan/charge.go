package xinyan

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var (
	charge = []float64{
		0.625,
		0.675625,
		0.726875,
		0.799375,
		0.85,
		0.908125,
		0.988125,
		1.068125,
		1.148125,
		1.235625,
		1.3225,
		1.41,
		1.496875,
		1.496875,
		1.496875,
	}
	chargeFinal = []float64{
		1.13,
		1.22153,
		1.31419,
		1.44527,
		1.5368,
		1.64189,
		1.78653,
		1.93117,
		2.07581,
		2.23401,
		2.39108,
		2.54928,
		2.70635,
		2.70635,
		2.70635,
	}
)

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	hold := p["hold"]
	if hold > 0 {
		if hold > 330 { // 10s
			hold = 330
		}
		return c.ChargeHold(hold), nil
	}
	return c.ChargePress(), nil
}

func (c *char) ChargeHold(duration int) action.Info {
	c6Bonus := 0.0
	if c.Base.Cons >= 6 {
		c6Bonus = c.TotalDef(false) * 0.5
	}

	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Charged Attack Cyclic DMG",
		AttackTag:          attacks.AttackTagExtra,
		ICDTag:             attacks.ICDTagNormalAttack,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeBlunt,
		Element:            attributes.Physical,
		Durability:         25,
		HitlagHaltFrames:   0.03 * 60,
		HitlagFactor:       0.01,
		CanBeDefenseHalted: true,
		FlatDmg:            (c.TotalAtk() + c6Bonus) * charge[c.TalentLvlAttack()],
	}

	ai2 := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Charged Attack Final DMG",
		AttackTag:          attacks.AttackTagExtra,
		ICDTag:             attacks.ICDTagNormalAttack,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeBlunt,
		Element:            attributes.Physical,
		Durability:         25,
		HitlagHaltFrames:   0.03 * 60,
		HitlagFactor:       0.01,
		CanBeDefenseHalted: true,
		FlatDmg:            (c.TotalAtk() + c6Bonus) * chargeFinal[c.TalentLvlAttack()],
	}

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0}, 3)

	for i := 0; i <= duration; i += 29 {
		c.Core.Tasks.Add(func() {
			c.Core.QueueAttack(ai, ap, 0, 0)
		}, 83+i)
	}

	c.Core.QueueAttack(ai2, ap, duration+24, duration+24)

	return action.Info{
		Frames:          func(next action.Action) int { return duration + 35 },
		AnimationLength: duration + 35,
		CanQueueAfter:   duration + 35,
		State:           action.ChargeAttackState,
	}
}

func (c *char) ChargePress() action.Info {
	c6Bonus := 0.0
	if c.Base.Cons >= 6 {
		c6Bonus = c.TotalDef(false) * 0.5
	}

	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Charged Attack Final DMG",
		AttackTag:          attacks.AttackTagExtra,
		ICDTag:             attacks.ICDTagNormalAttack,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeBlunt,
		Element:            attributes.Physical,
		Durability:         25,
		HitlagHaltFrames:   0.15 * 60,
		HitlagFactor:       0.01,
		CanBeDefenseHalted: true,
		FlatDmg:            (c.TotalAtk() + c6Bonus) * chargeFinal[c.TalentLvlAttack()],
	}

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0}, 3)

	c.Core.QueueAttack(ai, ap, 24, 24)
	return action.Info{
		Frames:          func(next action.Action) int { return 35 },
		AnimationLength: 35,
		CanQueueAfter:   35,
		State:           action.ChargeAttackState,
	}
}
