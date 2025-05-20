package escoffier

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
	attackHitmarks        = [][]int{{6}, {13}, {30, 48}}
	attackHitlagHaltFrame = [][]float64{{0.12}, {0.12}, {0.06, 0.03}}
	attackDefHalt         = [][]bool{{true}, {true}, {false, true}}
	attackHitboxes        = [][]float64{{2}, {2}, {2, 3.2}}
	attackOffsets         = []float64{0.8, 0.8, 0.8}
	attackFanAngles       = []float64{220, 220, 220}
)

const normalHitNum = 3

func init() {
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0][0], 19) // N1 -> CA
	attackFrames[0][action.ActionDash] = 14

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1][0], 24) // N2 -> CA
	attackFrames[1][action.ActionDash] = 14

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2][1], 62) // N3 -> CA
	attackFrames[2][action.ActionDash] = 51
	attackFrames[2][action.ActionCharge] = 500 // N3 -> CA, TODO: this action is illegal; need better way to handle it
}

// Normal attack damage queue generator
// Very standard
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
		ap := combat.NewCircleHitOnTargetFanAngle(
			c.Core.Combat.Player(),
			geometry.Point{Y: attackOffsets[c.NormalCounter]},
			attackHitboxes[c.NormalCounter][0],
			attackFanAngles[c.NormalCounter],
		)
		if c.NormalCounter == 3 {
			ap = combat.NewBoxHitOnTarget(
				c.Core.Combat.Player(),
				geometry.Point{Y: attackOffsets[c.NormalCounter]},
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
