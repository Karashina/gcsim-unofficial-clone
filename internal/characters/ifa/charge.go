package ifa

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
)

var chargeFrames []int

const (
	chargeHitmark             = 41
	chargeTonicInitialHitmark = 12
	chargeTonicInterval       = 43
	chargemaxDur              = 589
)

func init() {
	chargeFrames = frames.InitAbilSlice(41)

	chargeFrames[action.ActionDash] = chargeHitmark
	chargeFrames[action.ActionJump] = chargeHitmark
	chargeFrames[action.ActionSwap] = chargeHitmark
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	if c.nightsoulState.HasBlessing() {
		return c.attackSkillHold(p), nil
	}
	travel, ok := p["travel"]
	if !ok {
		travel = 10
	}
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Charge Attack",
		AttackTag:  attacks.AttackTagExtra,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       charge[c.TalentLvlAttack()],
	}

	// skip CA windup if we're in NA
	windup := 0
	if c.Core.Player.CurrentState() == action.NormalAttackState {
		windup = 14
	}

	// TODO: Not sure of snapshot timing
	c.Core.QueueAttack(
		ai,
		// TODO: Find actual radius of CA projectile
		combat.NewCircleHit(
			c.Core.Combat.Player(),
			c.Core.Combat.PrimaryTarget(),
			nil,
			3,
		),
		chargeHitmark-windup,
		chargeHitmark-windup+travel,
	)

	return action.Info{
		Frames:          func(next action.Action) int { return chargeFrames[next] - windup },
		AnimationLength: chargeFrames[action.InvalidAction] - windup,
		CanQueueAfter:   chargeHitmark - windup,
		State:           action.ChargeAttackState,
	}, nil
}
