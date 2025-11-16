package flins

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var chargeFrames []int

const chargeHitmark = 31

func init() {
	chargeFrames = frames.InitAbilSlice(39) // CA -> Walk
	chargeFrames[action.ActionAttack] = 59
}

// Charged Attack - part of Pocztowy Demonspear
func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(skillKey) {
		return c.chargeE() // go to Electro-infused charge attack during Manifest Flame form
	}

	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Charge",
		AttackTag:          attacks.AttackTagExtra,
		ICDTag:             attacks.ICDTagExtraAttack,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeSlash,
		Element:            attributes.Physical,
		Durability:         25,
		Mult:               charge[c.TalentLvlAttack()],
		HitlagHaltFrames:   0.09 * 60,
		HitlagFactor:       0.01,
		CanBeDefenseHalted: true,
	}

	c.Core.QueueAttack(
		ai,
		combat.NewBoxHitOnTarget(c.Core.Combat.PrimaryTarget(), geometry.Point{Y: -1.2}, 3.3, 3.5),
		chargeHitmark,
		chargeHitmark,
	)

	return action.Info{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeHitmark,
		State:           action.ChargeAttackState,
	}, nil
}

// Charged Attack during Manifest Flame form (Electro-infused)
// Ancient Rite: Arcane Light
func (c *char) chargeE() (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Ancient Rite: Arcane Light Charge",
		AttackTag:          attacks.AttackTagExtra,
		ICDTag:             attacks.ICDTagExtraAttack,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeSlash,
		Element:            attributes.Electro,
		Durability:         25,
		Mult:               charge_e[c.TalentLvlSkill()],
		HitlagHaltFrames:   0.00,
		HitlagFactor:       0.01,
		CanBeDefenseHalted: true,
		IgnoreInfusion:     true,
	}

	c.Core.QueueAttack(
		ai,
		combat.NewBoxHitOnTarget(c.Core.Combat.PrimaryTarget(), geometry.Point{Y: -1.2}, 3.3, 3.5),
		chargeHitmark,
		chargeHitmark,
		c.particleCB,
	)

	return action.Info{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeHitmark,
		State:           action.ChargeAttackState,
	}, nil
}

