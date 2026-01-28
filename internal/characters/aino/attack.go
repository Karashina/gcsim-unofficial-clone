package aino

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
	attackHitmarks        = [][]int{{25}, {22}, {34, 43}}
	attackPoiseDMG        = []float64{75, 75, 75}
	attackHitlagHaltFrame = [][]float64{{0.06}, {0.12}, {0, 0}}
	attackHitlagDefHalt   = [][]bool{{true}, {true}, {false, false}}
	attackOffsets         = []float64{0.5, 0.5, 1.0}
)

const normalHitNum = 3

func init() {
	attackFrames = make([][]int, normalHitNum)
	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0][0], 43)
	attackFrames[0][action.ActionDash] = 27

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1][0], 50)
	attackFrames[1][action.ActionDash] = 22

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2][1], 63)
	attackFrames[2][action.ActionDash] = 40
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	for i, mult := range attack[c.NormalCounter] {
		ai := combat.AttackInfo{
			ActorIndex:         c.Index,
			Abil:               fmt.Sprintf("Normal %v", c.NormalCounter),
			AttackTag:          attacks.AttackTagNormal,
			ICDTag:             attacks.ICDTagNormalAttack,
			ICDGroup:           attacks.ICDGroupDefault,
			StrikeType:         attacks.StrikeTypeBlunt,
			PoiseDMG:           attackPoiseDMG[c.NormalCounter],
			Element:            attributes.Physical,
			Durability:         25,
			Mult:               mult[c.TalentLvlAttack()],
			HitlagFactor:       0.01,
			HitlagHaltFrames:   attackHitlagHaltFrame[c.NormalCounter][i] * 60,
			CanBeDefenseHalted: attackHitlagDefHalt[c.NormalCounter][i],
		}
		ap := combat.NewCircleHitOnTarget(
			c.Core.Combat.Player(),
			geometry.Point{Y: attackOffsets[c.NormalCounter]},
			2,
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
