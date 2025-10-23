package nefer

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
)

var (
	attackFrames          [][]int
	attackHitmarks        = [][]int{{15}, {12}, {20}, {25}}
	attackHitlagFactor    = [][]float64{{0}, {0}, {0.01}, {0.01}}
	attackHitlagHaltFrame = [][]float64{{0}, {0}, {0.03}, {0.06}}
	attackHitboxes        = [][]float64{{2.5}, {2.5}, {3.0}, {3.5}}
	attackOffsets         = []float64{0, 0, 0, 0}
)

const normalHitNum = 4

func init() {
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0][0], 30)
	attackFrames[0][action.ActionAttack] = 17
	attackFrames[0][action.ActionCharge] = 25

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1][0], 28)
	attackFrames[1][action.ActionAttack] = 15
	attackFrames[1][action.ActionCharge] = 22

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2][0], 45)
	attackFrames[2][action.ActionAttack] = 26
	attackFrames[2][action.ActionCharge] = 22

	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmarks[3][0], 65)
	attackFrames[3][action.ActionCharge] = 40
	attackFrames[3][action.ActionWalk] = 60
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	for i, mult := range attack[c.NormalCounter] {
		ai := combat.AttackInfo{
			ActorIndex:         c.Index,
			Abil:               fmt.Sprintf("Normal %v", c.NormalCounter),
			AttackTag:          attacks.AttackTagNormal,
			ICDTag:             attacks.ICDTagNormalAttack,
			ICDGroup:           attacks.ICDGroupDefault,
			StrikeType:         attacks.StrikeTypeDefault,
			Element:            attributes.Dendro,
			Durability:         25,
			Mult:               mult[c.TalentLvlAttack()],
			HitlagFactor:       attackHitlagFactor[c.NormalCounter][i],
			HitlagHaltFrames:   attackHitlagHaltFrame[c.NormalCounter][i] * 60,
			CanBeDefenseHalted: true,
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
