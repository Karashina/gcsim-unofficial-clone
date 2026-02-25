package illuga

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
	attackHitmarks        = [][]int{{13}, {7}, {20, 36}, {43}}
	attackHitlagHaltFrame = [][]float64{{0.06}, {0.06}, {0.06, 0.00}, {0.1}}
	attackDefHalt         = [][]bool{{true}, {true}, {false, true}, {true}}
	attackHitboxes        = [][]float64{{2.0}, {2.0}, {2.0, 2.0}, {2.5}}
	attackOffsets         = []float64{0.8, 0.8, 0.8, 1.0}
)

const normalHitNum = 4

func init() {
	attackFrames = make([][]int, normalHitNum)

	// N1 -> x
	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0][0], 20)
	attackFrames[0][action.ActionAttack] = 30

	// N2 -> x
	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1][0], 13)
	attackFrames[1][action.ActionAttack] = 18

	// N3 -> x (2 hits)
	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2][1], 42)
	attackFrames[2][action.ActionAttack] = 47

	// N4 -> x
	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmarks[3][0], 53)
	attackFrames[3][action.ActionCharge] = 76
}

// Attack performs normal attacks (Physical, based on Thoma's polearm style)
func (c *char) Attack(p map[string]int) (action.Info, error) {
	for i, mult := range attack[c.NormalCounter] {
		ai := combat.AttackInfo{
			ActorIndex:         c.Index,
			Abil:               fmt.Sprintf("Normal %v", c.NormalCounter),
			AttackTag:          attacks.AttackTagNormal,
			ICDTag:             attacks.ICDTagNormalAttack,
			ICDGroup:           attacks.ICDGroupDefault,
			StrikeType:         attacks.StrikeTypeSpear,
			Element:            attributes.Physical,
			Durability:         25,
			Mult:               mult[c.TalentLvlAttack()],
			HitlagFactor:       0.01,
			HitlagHaltFrames:   attackHitlagHaltFrame[c.NormalCounter][i] * 60,
			CanBeDefenseHalted: attackDefHalt[c.NormalCounter][i],
		}

		ap := combat.NewCircleHitOnTarget(
			c.Core.Combat.Player(),
			geometry.Point{Y: attackOffsets[c.NormalCounter]},
			attackHitboxes[c.NormalCounter][0],
		)

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
