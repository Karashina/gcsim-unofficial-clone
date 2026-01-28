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
	attackFrames          [][]int
	attackHitmarks        = [][]int{{17}, {16}, {14, 33}, {21}}
	attackHitlagHaltFrame = [][]float64{{0.12}, {0.12}, {0.00, 0.06}, {0.06}}
	attackHitboxes        = [][]float64{{2}, {1.8, 3.8}, {2}, {2.1}}
	attackOffsets         = [][]float64{{0.6}, {-0.3}, {0.4, 0.9}, {1}}
)

const normalHitNum = 4

func init() {
	// NA cancels
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0][0], 23) // N1 -> Walk
	attackFrames[0][action.ActionAttack] = 22

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1][0], 26) // N2 -> Walk
	attackFrames[1][action.ActionAttack] = 30

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2][1], 33) // N3 -> Walk
	attackFrames[2][action.ActionAttack] = 45

	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmarks[3][0], 29) // N4 -> Walk
	attackFrames[3][action.ActionAttack] = 80
	attackFrames[3][action.ActionCharge] = 500 //TODO: this action is illegal; need better way to handle it
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	for i, mult := range attack[c.NormalCounter] {
		ai := combat.AttackInfo{
			ActorIndex:         c.Index,
			Abil:               fmt.Sprintf("Normal %v", c.NormalCounter),
			AttackTag:          attacks.AttackTagNormal,
			ICDTag:             attacks.ICDTagNormalAttack,
			ICDGroup:           attacks.ICDGroupDefault,
			StrikeType:         attacks.StrikeTypeSlash,
			Element:            attributes.Physical,
			Durability:         25,
			Mult:               mult[c.TalentLvlAttack()],
			HitlagFactor:       0.01,
			HitlagHaltFrames:   attackHitlagHaltFrame[c.NormalCounter][i] * 60,
			CanBeDefenseHalted: true,
		}
		ap := combat.NewCircleHitOnTarget(
			c.Core.Combat.Player(),
			geometry.Point{Y: attackOffsets[c.NormalCounter][i]},
			attackHitboxes[c.NormalCounter][0],
		)
		if c.NormalCounter == 1 {
			ap = combat.NewBoxHitOnTarget(
				c.Core.Combat.Player(),
				geometry.Point{Y: attackOffsets[c.NormalCounter][i]},
				attackHitboxes[c.NormalCounter][0],
				attackHitboxes[c.NormalCounter][1],
			)
		}
		c.QueueCharTask(func() {
			c.Core.QueueAttack(ai, ap, 0, 0)
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
