package mualani

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
)

var chargeFramesNormal []int

const chargeHitmarkNormal = 70

func init() {
	chargeFramesNormal = frames.InitAbilSlice(74)
	chargeFramesNormal[action.ActionAttack] = 74
	chargeFramesNormal[action.ActionCharge] = 74
	chargeFramesNormal[action.ActionSkill] = 74
	chargeFramesNormal[action.ActionBurst] = 74
	chargeFramesNormal[action.ActionDash] = 70
	chargeFramesNormal[action.ActionJump] = 70
	chargeFramesNormal[action.ActionSwap] = 70
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Charge Attack",
		AttackTag:  attacks.AttackTagExtra,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
		Mult:       charge[c.TalentLvlAttack()],
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3),
		chargeHitmarkNormal,
		chargeHitmarkNormal,
	)
	atkspd := c.Stat(attributes.AtkSpd)
	return action.Info{
		Frames: func(next action.Action) int {
			return frames.AtkSpdAdjust(chargeFramesNormal[next], atkspd)
		},
		AnimationLength: chargeFramesNormal[action.InvalidAction],
		CanQueueAfter:   chargeFramesNormal[action.ActionDash],
		State:           action.ChargeAttackState,
	}, nil
}
