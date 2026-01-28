package nefer

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
	attackFrames   [][]int
	attackHitmarks = [][]int{{14}, {13}, {8, 26}, {6}}
	attackHitboxes = [][]float64{{2.5}, {2.5}, {3.0}, {3.5}}
	attackOffsets  = []float64{0, 0, 0, 0}
)

const normalHitNum = 4

func init() {
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0][0], 29)
	attackFrames[0][action.ActionDash] = attackHitmarks[0][0]

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1][0], 25)
	attackFrames[1][action.ActionDash] = attackHitmarks[1][0]

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2][1], 63)
	attackFrames[2][action.ActionDash] = attackHitmarks[2][1]

	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmarks[3][0], 41)
	attackFrames[3][action.ActionDash] = attackHitmarks[3][0]
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	for i, mult := range attack[c.NormalCounter] {
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       fmt.Sprintf("Normal %v", c.NormalCounter),
			AttackTag:  attacks.AttackTagNormal,
			ICDTag:     attacks.ICDTagNormalAttack,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Dendro,
			Durability: 25,
			Mult:       mult[c.TalentLvlAttack()],
		}

		c.QueueCharTask(func() {
			c.Core.QueueAttack(
				ai,
				combat.NewCircleHitOnTarget(
					c.Core.Combat.Player(),
					geometry.Point{Y: attackOffsets[c.NormalCounter]},
					attackHitboxes[c.NormalCounter][0],
				),
				0,
				0,
			)
		}, attackHitmarks[c.NormalCounter][i])
	}

	defer c.AdvanceNormalIndex()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackFrames[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[c.NormalCounter][len(attackHitmarks[c.NormalCounter])-1],
		State:           action.NormalAttackState,
	}, nil
}
