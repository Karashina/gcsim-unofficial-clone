package lanyan

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
	attackHits            = []int{1, 2, 2, 1}
	attackHitmarks        = [][]int{{12}, {16, 34}, {15, 16}, {40}}
	attackHitlagHaltFrame = [][]float64{{0}, {0.03, 0}, {0.03, 0}, {0.06}}
	attackDefHalt         = [][]bool{{false}, {true, true}, {true, true}, {true}}
	attackHitboxes        = []float64{2, 2, 2, 2}
	attackOffsets         = []float64{-0.2, -0.2, -0.2, -0.2}
)

const normalHitNum = 4

func init() {
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0][0], 27)
	attackFrames[0][action.ActionAttack] = 27
	attackFrames[0][action.ActionCharge] = 21
	attackFrames[0][action.ActionDash] = 18
	attackFrames[0][action.ActionJump] = 23

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1][1], 52)
	attackFrames[1][action.ActionAttack] = 52
	attackFrames[1][action.ActionCharge] = 46
	attackFrames[1][action.ActionDash] = 37
	attackFrames[1][action.ActionJump] = 42

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2][1], 45)
	attackFrames[2][action.ActionAttack] = 45
	attackFrames[2][action.ActionCharge] = 45
	attackFrames[2][action.ActionDash] = 30
	attackFrames[2][action.ActionJump] = 35

	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmarks[3][0], 63)
	attackFrames[3][action.ActionAttack] = 63
	attackFrames[3][action.ActionDash] = 21
	attackFrames[3][action.ActionJump] = 49
	attackFrames[3][action.ActionCharge] = 58
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(SkillKey) {
		return c.feathermoonRing()
	}
	for i := 0; i < attackHits[c.NormalCounter]; i++ {
		ai := combat.AttackInfo{
			ActorIndex:         c.Index,
			Abil:               fmt.Sprintf("Normal %v", c.NormalCounter),
			AttackTag:          attacks.AttackTagNormal,
			ICDTag:             attacks.ICDTagNormalAttack,
			ICDGroup:           attacks.ICDGroupDefault,
			StrikeType:         attacks.StrikeTypeDefault,
			Element:            attributes.Anemo,
			Durability:         25,
			Mult:               attack[c.NormalCounter][i][c.TalentLvlAttack()],
			HitlagFactor:       0.01,
			HitlagHaltFrames:   attackHitlagHaltFrame[c.NormalCounter][i] * 60,
			CanBeDefenseHalted: attackDefHalt[c.NormalCounter][i],
		}
		ap := combat.NewCircleHitOnTarget(
			c.Core.Combat.Player(),
			geometry.Point{Y: attackOffsets[c.NormalCounter]},
			attackHitboxes[c.NormalCounter],
		)
		c.Core.QueueAttack(ai, ap, attackHitmarks[c.NormalCounter][i], attackHitmarks[c.NormalCounter][i])
	}

	defer c.AdvanceNormalIndex()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackFrames[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[c.NormalCounter][len(attackHitmarks[c.NormalCounter])-1],
		State:           action.NormalAttackState,
	}, nil
}
