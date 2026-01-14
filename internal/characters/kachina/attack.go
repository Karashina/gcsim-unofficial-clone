package kachina

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
	attackHitmarks        = [][]int{{15}, {9, 35}, {21}, {16}}
	attackHitlagHaltFrame = [][]float64{{0.03}, {0, 0.03}, {0.03}, {0.04}}
	attackDefHalt         = [][]bool{{true}, {false, true}, {false}, {true}}
	attackHitboxes        = []float64{2, 2, 2, 2.5}
	attackOffsets         = []float64{0.8, 0.8, 0.8, 1.1}
	attackFanAngles       = []float64{360, 220, 220, 90}
)

const (
	normalHitNum  = 4
	attackFramesE = 35
)

func init() {
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0][0], 32)
	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1][1], 52)
	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2][0], 46)
	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmarks[3][0], 56)
	attackFrames[3][action.ActionCharge] = 500 // N5 -> CA, TODO: this action is illegal; need better way to handle it
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(skillRideKey) {
		return c.AttackRide(32)
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
		ap := combat.NewCircleHitOnTargetFanAngle(
			c.Core.Combat.Player(),
			geometry.Point{Y: attackOffsets[c.NormalCounter]},
			attackHitboxes[c.NormalCounter],
			attackFanAngles[c.NormalCounter],
		)
		if c.NormalCounter == 3 {
			ap = combat.NewBoxHitOnTarget(
				c.Core.Combat.Player(),
				geometry.Point{Y: attackOffsets[c.NormalCounter]},
				attackHitboxes[c.NormalCounter],
				attackHitboxes[c.NormalCounter],
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

func (c *char) AttackRide(hitmark int) (action.Info, error) {

	c.QueueCharTask(c.TwirlyRideAttack(), hitmark)

	return action.Info{
		Frames:          func(next action.Action) int { return attackFramesE },
		AnimationLength: attackFramesE,
		CanQueueAfter:   attackFramesE,
		State:           action.NormalAttackState,
	}, nil
}

