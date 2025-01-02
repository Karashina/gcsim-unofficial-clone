package mavuika

import (
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
)

const chargeKey = "mavuika-charge"

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	if c.nightsoulState.HasBlessing() {
		hold := p["hold"]
		if hold > 0 {
			if hold > 360 { // 10s
				hold = 360
			}
			return c.ChargeBike(hold)
		}
		return c.ChargeBike(hold)
	} else {
		return c.ChargePress()
	}
}

func (c *char) ChargeBike(duration int) (action.Info, error) {
	buff := 0.0
	if c.StatusIsActive(BurstKey) {
		buff = c.consumedspirit
	}
	c2buff := 0.0
	if c.Base.Cons >= 2 {
		c2buff = 0.9
	}

	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Flamestrider Charged Attack Cyclic DMG",
		AttackTag:          attacks.AttackTagExtra,
		AdditionalTags:     []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:             attacks.ICDTagNormalAttack,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeBlunt,
		Element:            attributes.Pyro,
		Durability:         25,
		HitlagHaltFrames:   0.03 * 60,
		HitlagFactor:       0.01,
		CanBeDefenseHalted: true,
		Mult:               bikecharge[c.TalentLvlSkill()] + burstnabonus[c.TalentLvlBurst()]*buff + c2buff,
		IgnoreInfusion:     true,
	}

	ai2 := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Flamestrider Charged Attack Final DMG",
		AttackTag:          attacks.AttackTagExtra,
		AdditionalTags:     []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:             attacks.ICDTagNormalAttack,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeBlunt,
		Element:            attributes.Pyro,
		Durability:         25,
		HitlagHaltFrames:   0.03 * 60,
		HitlagFactor:       0.01,
		CanBeDefenseHalted: true,
		Mult:               bikechargefinal[c.TalentLvlSkill()] + burstcabonus[c.TalentLvlBurst()]*buff + c2buff,
		IgnoreInfusion:     true,
	}

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0}, 3)

	for i := 0; i <= duration; i += 41 {
		c.Core.Tasks.Add(func() {
			c.Core.QueueAttack(ai, ap, 0, 0)
		}, 83+i)
	}

	c.Core.QueueAttack(ai2, ap, duration+43, duration+43)
	c.AddStatus(chargeKey, duration+57, true)

	return action.Info{
		Frames:          func(next action.Action) int { return duration + 57 },
		AnimationLength: duration + 57,
		CanQueueAfter:   duration + 57,
		State:           action.ChargeAttackState,
	}, nil
}

func (c *char) ChargePress() (action.Info, error) {

	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Charged Attack",
		AttackTag:          attacks.AttackTagExtra,
		ICDTag:             attacks.ICDTagNormalAttack,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeBlunt,
		Element:            attributes.Physical,
		Durability:         25,
		HitlagHaltFrames:   0.15 * 60,
		HitlagFactor:       0.01,
		CanBeDefenseHalted: true,
		Mult:               charge[c.TalentLvlAttack()],
	}

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0}, 4)

	c.Core.QueueAttack(ai, ap, 60, 60)
	return action.Info{
		Frames:          func(next action.Action) int { return 95 },
		AnimationLength: 95,
		CanQueueAfter:   95,
		State:           action.ChargeAttackState,
	}, nil
}
