package dhalia

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var (
	chargeFrames   []int
	chargeHitmarks = []int{11, 20}
	chargeOffsets  = []float64{1, 1}
)

func init() {
	chargeFrames = frames.InitAbilSlice(27) // C -> Walk
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	for i, mult := range charge {
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       fmt.Sprintf("Charge %v", i),
			AttackTag:  attacks.AttackTagExtra,
			ICDTag:     attacks.ICDTagNormalAttack,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeSlash,
			Element:    attributes.Physical,
			Durability: 25,
			Mult:       mult[c.TalentLvlAttack()],
		}
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(
				c.Core.Combat.Player(),
				geometry.Point{Y: chargeOffsets[i]},
				2.8,
			),
			chargeHitmarks[i],
			chargeHitmarks[i],
		)
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeHitmarks[len(chargeHitmarks)-1],
		State:           action.ChargeAttackState,
	}, nil
}
