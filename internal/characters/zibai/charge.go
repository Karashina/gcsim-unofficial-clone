package zibai

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
	chargeFrames = frames.InitAbilSlice(80)
	chargeFrames[action.ActionDash] = 47
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	// Check if in Lunar Phase Shift mode
	if c.lunarPhaseShiftActive {
		return c.lunarPhaseShiftCharge(p)
	}

	// Normal charged attack (Physical) - 2 hits
	mults := [][]float64{charge_1, charge_2}
	for i, mult := range mults {
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Charged Attack",
			AttackTag:  attacks.AttackTagExtra,
			ICDTag:     attacks.ICDTagNormalAttack,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeSlash,
			Element:    attributes.Physical,
			Durability: 25,
			Mult:       mult[c.TalentLvlAttack()],
		}

		ap := combat.NewBoxHitOnTarget(
			c.Core.Combat.Player(),
			geometry.Point{Y: 0.8},
			2.5,
			4.0,
		)

		delay := chargeHitmark + i*8

		c.QueueCharTask(func() {
			c.Core.QueueAttack(ai, ap, 0, 0)
		}, delay)
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeHitmark,
		State:           action.ChargeAttackState,
	}, nil
}

// lunarPhaseShiftCharge handles charged attacks during Lunar Phase Shift mode (Geo, DEF-scaling)
func (c *char) lunarPhaseShiftCharge(p map[string]int) (action.Info, error) {
	// Lunar Phase Shift charged attack (Geo, DEF-scaling) - 2 hits
	mults := [][]float64{lunarPhaseShiftCharge_1, lunarPhaseShiftCharge_2}
	for i, mult := range mults {
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Lunar Phase Shift Charged Attack",
			AttackTag:  attacks.AttackTagExtra,
			ICDTag:     attacks.ICDTagNormalAttack,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeSlash,
			Element:    attributes.Geo,
			Durability: 25,
			UseDef:     true,
			Mult:       mult[c.TalentLvlSkill()],
		}

		ap := combat.NewBoxHitOnTarget(
			c.Core.Combat.Player(),
			geometry.Point{Y: 0.8},
			2.5,
			4.0,
		)

		delay := chargeHitmark + i*8

		c.QueueCharTask(func() {
			c.Core.QueueAttack(ai, ap, 0, 0, c.radianceGainCB)
		}, delay)
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeHitmark,
		State:           action.ChargeAttackState,
	}, nil
}
