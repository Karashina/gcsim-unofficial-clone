package illuga

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var chargeFrames []int

const chargeHitmark = 26

func init() {
	chargeFrames = frames.InitAbilSlice(67)
	chargeFrames[action.ActionDash] = 26
}

// ChargeAttack performs charged attack (Physical, lunge forward)
func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Charged Attack",
		AttackTag:  attacks.AttackTagExtra,
		ICDTag:     attacks.ICDTagNormalAttack,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeSpear,
		Element:    attributes.Physical,
		Durability: 25,
		Mult:       charge[c.TalentLvlAttack()],
	}

	ap := combat.NewBoxHitOnTarget(
		c.Core.Combat.Player(),
		geometry.Point{Y: 1.0},
		2.0,
		5.0,
	)

	c.QueueCharTask(func() {
		c.Core.QueueAttack(ai, ap, 0, 0)
	}, chargeHitmark)

	return action.Info{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeHitmark,
		State:           action.ChargeAttackState,
	}, nil
}
