package ineffa

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
	attackHitmarks        = [][]int{{10}, {15}, {7, 16}, {28}}
	attackHitlagHaltFrame = [][]float64{{0.12}, {0.12}, {0.03, 0.12}, {0.10}}
	attackDefHalt         = [][]bool{{true}, {true}, {false, true}, {true}}
	attackHitboxes        = [][][]float64{{{1.6, 2}}, {{2}}, {{2.5}, {2.5}}, {{2.5}}}
	attackOffsets         = [][]float64{{1}, {-0.2}, {-0.2, -0.2}, {0.2}}
)

const normalHitNum = 4

func init() {
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0][0], 22)
	attackFrames[0][action.ActionCharge] = 35
	attackFrames[0][action.ActionDash] = 10

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1][0], 30)
	attackFrames[1][action.ActionDash] = 20

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2][1], 36)
	attackFrames[2][action.ActionDash] = 22

	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmarks[3][0], 57)
	attackFrames[3][action.ActionDash] = 37
	attackFrames[3][action.ActionCharge] = 500 // Illegal action; needs better handling
}

// Normal attack implementation
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
			CanBeDefenseHalted: attackDefHalt[c.NormalCounter][i],
		}

		ap := combat.NewCircleHitOnTarget(
			c.Core.Combat.Player(),
			geometry.Point{Y: attackOffsets[c.NormalCounter][i]},
			attackHitboxes[c.NormalCounter][i][0],
		)
		if c.NormalCounter == 0 {
			ai.StrikeType = attacks.StrikeTypeSpear
			ap = combat.NewBoxHitOnTarget(
				c.Core.Combat.Player(),
				geometry.Point{Y: attackOffsets[c.NormalCounter][i]},
				attackHitboxes[c.NormalCounter][i][0],
				attackHitboxes[c.NormalCounter][i][1],
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
