package lauma

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
)

var chargeFrames []int

const chargeHitmark = 35

func init() {
	chargeFrames = frames.InitAbilSlice(70) // total duration
	chargeFrames[action.ActionDash] = chargeHitmark
	chargeFrames[action.ActionJump] = chargeHitmark
	chargeFrames[action.ActionSwap] = chargeHitmark
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Charged Attack",
		AttackTag:  attacks.AttackTagExtra,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Pyro,
		Durability: 25,
		Mult:       charge[c.TalentLvlAttack()],
	}

	// Main charged attack
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 2}, 1.5),
		chargeHitmark,
		chargeHitmark,
	)

	// A1 passive: If HP > 50%, charged attack creates additional smaller explosions
	if c.Base.Ascension >= 1 && c.CurrentHPRatio() > 0.5 {
		for i := 0; i < 3; i++ {
			c.Core.Tasks.Add(func() {
				ai2 := ai
				ai2.Abil = "Charged Attack (A1)"
				ai2.Mult = charge[c.TalentLvlAttack()] * 0.3 // 30% of main damage
				
				c.Core.QueueAttack(
					ai2,
					combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 2 + float64(i)}, 1.0),
					0,
					0,
				)
			}, chargeHitmark + 10 + i*5) // Staggered explosions
		}
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeFrames[action.ActionDash],
		State:           action.ChargeAttackState,
	}, nil
}

// charge[TalentLevel-1]
var charge = []float64{1.568, 1.696, 1.824, 1.96, 2.088, 2.216, 2.352, 2.488, 2.624, 2.76, 2.896}