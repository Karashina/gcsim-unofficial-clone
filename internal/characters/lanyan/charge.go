package lanyan

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
)

var (
	chargeFrames   []int
	chargeRelease  = 40
	chargeHitmarks = []int{33, 42, 50}
)

func init() {
	chargeFrames = frames.InitAbilSlice(42) // C > NA
	chargeFrames[action.ActionAttack] = 42
	chargeFrames[action.ActionJump] = 38
	chargeFrames[action.ActionDash] = 34
	chargeFrames[action.ActionCharge] = 500 // Illegal action
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	travel, ok := p["travel"]
	if !ok {
		travel = 10
	}
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Charge Attack",
		AttackTag:  attacks.AttackTagExtra,
		ICDTag:     attacks.ICDTagExtraAttack,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       charge[c.TalentLvlAttack()],
	}
	for i := 0; i < 3; i++ {
		c.Core.QueueAttack(ai, combat.NewCircleHit(c.Core.Combat.Player(), c.Core.Combat.PrimaryTarget(), nil, 3), chargeRelease, chargeHitmarks[i]+travel)
	}
	return action.Info{
		Frames:          func(next action.Action) int { return chargeFrames[next] },
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeRelease,
		State:           action.ChargeAttackState,
	}, nil
}
