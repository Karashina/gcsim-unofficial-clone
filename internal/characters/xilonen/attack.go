package xilonen

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
	attackFrames           [][]int
	attackHitmarks         = [][]int{{19}, {13, 35}, {24}}
	attackHitlagHaltFrames = []float64{0.03, 0.03, 0.06}
	attackHitboxes         = []float64{0.5, 1.5, 2}
	attackOffsets          = [][]float64{{0.8}, {0.6, 0.6}, {0}}

	attackFramesBR          [][]int
	attackHitmarksBR        = []int{19, 16, 24, 31}
	attackHitlagHaltFrameBR = []float64{0.03, 0.03, 0.03, 0.06}
	attackRadiusBR          = []float64{3, 3, 2, 4}
	attackOffsetsBR         = []float64{0, 0, 2, 0}
	attackFanAnglesBR       = []float64{270, 180, 90, 270}
)

const normalHitNum = 3
const normalHitNumBR = 4

func init() {
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0][0], 30)
	attackFrames[0][action.ActionAttack] = 30

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1][1], 52)
	attackFrames[1][action.ActionAttack] = 52

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2][0], 61)
	attackFrames[2][action.ActionCharge] = 500 // Illegal action

	attackFramesBR = make([][]int, normalHitNumBR)

	attackFramesBR[0] = frames.InitNormalCancelSlice(attackHitmarksBR[0], 27)
	attackFramesBR[0][action.ActionAttack] = 27

	attackFramesBR[1] = frames.InitNormalCancelSlice(attackHitmarksBR[1], 30)
	attackFramesBR[1][action.ActionAttack] = 30

	attackFramesBR[2] = frames.InitNormalCancelSlice(attackHitmarksBR[2], 36)
	attackFramesBR[2][action.ActionAttack] = 36

	attackFramesBR[3] = frames.InitNormalCancelSlice(attackHitmarksBR[3], 72)
	attackFramesBR[3][action.ActionAttack] = 72
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(skillKey) {
		return c.BladeRoller(p)
	}

	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		AttackTag:          attacks.AttackTagNormal,
		ICDTag:             attacks.ICDTagNormalAttack,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeSlash,
		Element:            attributes.Physical,
		Durability:         25,
		HitlagHaltFrames:   attackHitlagHaltFrames[c.NormalCounter] * 60,
		HitlagFactor:       0.01,
		CanBeDefenseHalted: true,
	}

	for i, mult := range attack[c.NormalCounter] {
		ax := ai
		ax.Abil = fmt.Sprintf("Normal %v", c.NormalCounter)
		ax.Mult = mult[c.TalentLvlAttack()]
		ap := combat.NewCircleHitOnTarget(
			c.Core.Combat.Player(),
			geometry.Point{Y: attackOffsets[c.NormalCounter][i]},
			attackHitboxes[c.NormalCounter],
		)
		c.QueueCharTask(func() {
			c.Core.QueueAttack(ax, ap, 0, 0)
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

func (c *char) BladeRoller(p map[string]int) (action.Info, error) {

	ai := combat.AttackInfo{
		ActorIndex:       c.Index,
		Abil:             fmt.Sprintf("Normal: Blade Roller %v", c.NormalCounter),
		AttackTag:        attacks.AttackTagNormal,
		ICDTag:           attacks.ICDTagNormalAttack,
		ICDGroup:         attacks.ICDGroupDefault,
		StrikeType:       attacks.StrikeTypeSlash,
		Element:          attributes.Geo,
		Durability:       25,
		Mult:             bladeroller[c.NormalCounter][c.TalentLvlAttack()],
		HitlagFactor:     0.01,
		HitlagHaltFrames: attackHitlagHaltFrameBR[c.NormalCounter] * 60,
		UseDef:           true,
		Alignment:        attacks.AdditionalTagNightsoul,
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTargetFanAngle(
			c.Core.Combat.Player(),
			geometry.Point{Y: attackOffsetsBR[c.NormalCounter]},
			attackRadiusBR[c.NormalCounter],
			attackFanAnglesBR[c.NormalCounter],
		),
		attackHitmarksBR[c.NormalCounter],
		attackHitmarksBR[c.NormalCounter],
	)

	defer c.AdvanceNormalIndex()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFramesBR),
		AnimationLength: attackFramesBR[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackHitmarksBR[c.NormalCounter],
		State:           action.NormalAttackState,
	}, nil
}
