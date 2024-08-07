package emilie

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
)

var (
	attackFrames          [][]int
	attackHitmarks        = [][]int{{15}, {18}, {28}, {35}}
	attackHitlagHaltFrame = [][]float64{{0.03}, {0.03}, {0.03}, {0.09}}
	attackHitboxes        = [][]float64{
		{2, 2},
		{1.5, 3.5},
		{5, 3},
		{2, 2},
	}
)

const (
	normalHitNum = 4
)

func init() {
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0][0], 36)
	attackFrames[0][action.ActionCharge] = 34

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1][0], 36)
	attackFrames[1][action.ActionCharge] = 38

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2][0], 35)
	attackFrames[2][action.ActionCharge] = 40

	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmarks[3][0], 77)
	attackFrames[3][action.ActionCharge] = 500 //TODO: this action is illegal; need better way to handle it
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	for i, mult := range attack[c.NormalCounter] {
		ai := combat.AttackInfo{
			ActorIndex:         c.Index,
			Abil:               fmt.Sprintf("Normal %v", c.NormalCounter),
			Mult:               mult[c.TalentLvlAttack()],
			AttackTag:          attacks.AttackTagNormal,
			ICDTag:             attacks.ICDTagNormalAttack,
			ICDGroup:           attacks.ICDGroupDefault,
			StrikeType:         attacks.StrikeTypeSpear,
			Element:            attributes.Physical,
			Durability:         25,
			HitlagFactor:       0.01,
			HitlagHaltFrames:   attackHitlagHaltFrame[c.NormalCounter][i] * 60,
			CanBeDefenseHalted: true,
		}
		if c.Base.Cons >= 6 && c.StatusIsActive(c6Key) {
			ai.Element = attributes.Dendro
			ai.FlatDmg += c.TotalAtk() * 3
			c.c6handle()
		}
		ap := combat.NewBoxHitOnTarget(
			c.Core.Combat.Player(),
			nil,
			attackHitboxes[c.NormalCounter][0],
			attackHitboxes[c.NormalCounter][1],
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
