package mizuki

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
)

var (
	attackFrames   [][]int
	attackHitmarks = []int{18, 26, 44}
	attackRadius   = []float64{1, 1, 1}
)

const normalHitNum = 3

func init() {
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0], 23)
	attackFrames[0][action.ActionAttack] = 23
	attackFrames[0][action.ActionCharge] = 23
	attackFrames[0][action.ActionSkill] = attackHitmarks[0]
	attackFrames[0][action.ActionBurst] = attackHitmarks[0]
	attackFrames[0][action.ActionDash] = attackHitmarks[0]
	attackFrames[0][action.ActionJump] = attackHitmarks[0]
	attackFrames[0][action.ActionSwap] = attackHitmarks[0]

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1], 33)
	attackFrames[1][action.ActionAttack] = 33
	attackFrames[1][action.ActionCharge] = 33
	attackFrames[1][action.ActionSkill] = attackHitmarks[1]
	attackFrames[1][action.ActionBurst] = attackHitmarks[1]
	attackFrames[1][action.ActionDash] = attackHitmarks[1]
	attackFrames[1][action.ActionJump] = attackHitmarks[1]
	attackFrames[1][action.ActionSwap] = attackHitmarks[1]

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2], 80)
	attackFrames[2][action.ActionAttack] = 80
	attackFrames[2][action.ActionCharge] = 80
	attackFrames[2][action.ActionSkill] = attackHitmarks[2]
	attackFrames[2][action.ActionBurst] = attackHitmarks[2]
	attackFrames[2][action.ActionDash] = attackHitmarks[2]
	attackFrames[2][action.ActionJump] = attackHitmarks[2]
	attackFrames[2][action.ActionSwap] = attackHitmarks[2]
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	travel, ok := p["travel"]
	if !ok {
		travel = 10
	}
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       fmt.Sprintf("Normal %v", c.NormalCounter),
		AttackTag:  attacks.AttackTagNormal,
		ICDTag:     attacks.ICDTagNormalAttack,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       attack[c.NormalCounter][c.TalentLvlAttack()],
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHit(
			c.Core.Combat.Player(),
			c.Core.Combat.PrimaryTarget(),
			nil,
			attackRadius[c.NormalCounter],
		),
		attackHitmarks[c.NormalCounter],
		attackHitmarks[c.NormalCounter]+travel,
	)

	defer c.AdvanceNormalIndex()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackFrames[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[c.NormalCounter],
		State:           action.NormalAttackState,
	}, nil
}
