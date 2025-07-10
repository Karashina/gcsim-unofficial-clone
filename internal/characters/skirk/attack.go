package skirk

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
	attackHitmarks        = [][]int{{10}, {8}, {8, 11}, {10}, {34}}
	attackHitlagHaltFrame = [][]float64{{0.06}, {0.06}, {0.06, 0.03}, {0.12}, {0.06}}
	attackDefHalt         = [][]bool{{true}, {true}, {true, true}, {true}, {true}}
	attackHitboxes        = []float64{1.7, 1.9, 2.1, 2.5, 2.5}
	attackOffsets         = []float64{1, 1, 1, 1, 1}

	spattackFrames          [][]int
	spattackHitmarks        = [][]int{{8}, {14}, {14, 28}, {11, 31}, {32}}
	spattackHitlagHaltFrame = [][]float64{{0.02}, {0.06}, {0.03, 0.02}, {0.06, 0.02}, {0.03}}
	spattackDefHalt         = [][]bool{{true}, {true}, {true, true}, {true, true}, {true}}
	spattackHitboxes        = []float64{1.7, 1.9, 2.1, 2.5, 2.5}
	spattackOffsets         = []float64{1, 1, 1, 1, 1}
)

const (
	normalHitNum   = 5
	spnormalHitNum = 5
)

func init() {
	// Normal attack
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0][0], 19)
	attackFrames[0][action.ActionAttack] = 29

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1][0], 17)
	attackFrames[1][action.ActionAttack] = 22

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2][1], 13)
	attackFrames[2][action.ActionAttack] = 40

	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmarks[3][0], 16)
	attackFrames[3][action.ActionAttack] = 20

	attackFrames[4] = frames.InitNormalCancelSlice(attackHitmarks[4][0], 36)
	attackFrames[4][action.ActionAttack] = 60

	// Normal attack
	spattackFrames = make([][]int, normalHitNum)

	spattackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0][0], 14)
	spattackFrames[0][action.ActionAttack] = 19

	spattackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1][0], 4)
	spattackFrames[1][action.ActionAttack] = 10

	spattackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2][1], 36)
	spattackFrames[2][action.ActionAttack] = 40

	spattackFrames[3] = frames.InitNormalCancelSlice(attackHitmarks[3][0], 36)
	spattackFrames[3][action.ActionAttack] = 36

	spattackFrames[4] = frames.InitNormalCancelSlice(attackHitmarks[4][0], 35)
	spattackFrames[4][action.ActionAttack] = 66
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	if c.onSevenPhaseFlash {
		return c.skillAttack(p)
	}

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

func (c *char) skillAttack(_ map[string]int) (action.Info, error) {

	for i, mult := range spattack[c.NormalCounter] {
		ai := combat.AttackInfo{
			ActorIndex:         c.Index,
			Abil:               fmt.Sprintf("Seven-Phase Flash Normal %v", c.NormalCounter),
			AttackTag:          attacks.AttackTagNormal,
			ICDTag:             attacks.ICDTagNormalAttack,
			ICDGroup:           attacks.ICDGroupDefault,
			StrikeType:         attacks.StrikeTypeSlash,
			Element:            attributes.Cryo,
			Durability:         25,
			Mult:               mult[c.TalentLvlSkill()] * c.a4BuffNA,
			HitlagFactor:       0.01,
			HitlagHaltFrames:   spattackHitlagHaltFrame[c.NormalCounter][i] * 60,
			CanBeDefenseHalted: spattackDefHalt[c.NormalCounter][i],
		}

		ap := combat.NewCircleHitOnTarget(
			c.Core.Combat.Player(),
			geometry.Point{Y: spattackOffsets[c.NormalCounter]},
			spattackHitboxes[c.NormalCounter],
		)

		c.Core.QueueAttack(ai, ap, spattackHitmarks[c.NormalCounter][i], spattackHitmarks[c.NormalCounter][i], c.particleCB)
		if c.NormalCounter == 2 || c.NormalCounter == 4 {
			c.c6("na")
		}
	}

	defer c.AdvanceNormalIndex()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, spattackFrames),
		AnimationLength: spattackFrames[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   spattackHitmarks[c.NormalCounter][len(spattackHitmarks[c.NormalCounter])-1],
		State:           action.NormalAttackState,
	}, nil
}
