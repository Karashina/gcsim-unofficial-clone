package varesa

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
	"github.com/genshinsim/gcsim/pkg/core/player"
)

var (
	chargeFrames    []int
	chargeFramesFP  []int
	ChargeHitmark   = 67
	ChargeHitmarkFP = 67
)

func init() {
	chargeFrames = frames.InitAbilSlice(146) // attack
	chargeFrames[action.ActionLowPlunge] = 84
	chargeFrames[action.ActionHighPlunge] = chargeFrames[action.ActionLowPlunge] + 1

	chargeFramesFP = frames.InitAbilSlice(146) // attack
	chargeFramesFP[action.ActionLowPlunge] = 84
	chargeFramesFP[action.ActionHighPlunge] = chargeFrames[action.ActionLowPlunge] + 1
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	if c.fastCharge {
		chargeFrames = frames.InitAbilSlice(30)
		chargeFramesFP = frames.InitAbilSlice(30)

		ChargeHitmark = 12
		ChargeHitmarkFP = 12

		c.fastCharge = false
	}
	if c.StatusIsActive(fieryPassionKey) {
		return c.ChargeFP(p)
	}

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Charge Attack",
		AttackTag:  attacks.AttackTagExtra,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       charge[c.TalentLvlAttack()],
	}
	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, 3.5)
	c.Core.QueueAttack(
		ai,
		ap,
		ChargeHitmark,
		ChargeHitmark,
	)

	c.Core.Player.SetAirborne(player.AirborneVaresa)

	return action.Info{
		Frames:          func(next action.Action) int { return chargeFrames[next] },
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeFrames[action.ActionDash],
		State:           action.ChargeAttackState,
	}, nil
}

//---------------------Fiery Passion--------------------------//

func (c *char) ChargeFP(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Charge Attack (Fiery Passion)",
		AttackTag:  attacks.AttackTagExtra,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       chargefp[c.TalentLvlAttack()],
	}
	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, 3.5)
	c.Core.QueueAttack(
		ai,
		ap,
		ChargeHitmarkFP,
		ChargeHitmarkFP,
	)

	c.Core.Player.SetAirborne(player.AirborneVaresa)

	return action.Info{
		Frames:          func(next action.Action) int { return chargeFramesFP[next] },
		AnimationLength: chargeFramesFP[action.InvalidAction],
		CanQueueAfter:   chargeFramesFP[action.ActionDash],
		State:           action.ChargeAttackState,
	}, nil
}
