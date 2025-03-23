package iansan

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
)

const normalHitNum = 3

var (
	attackFrames   [][]int
	attackHitmarks = []int{11, 9, 31}
)

func init() {
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0], 33) // walk
	attackFrames[0][action.ActionAttack] = 23
	attackFrames[0][action.ActionCharge] = 22
	attackFrames[0][action.ActionSkill] = 13
	attackFrames[0][action.ActionBurst] = 11
	attackFrames[0][action.ActionDash] = 12
	attackFrames[0][action.ActionJump] = 13
	attackFrames[0][action.ActionSwap] = 11

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1], 32) // walk
	attackFrames[1][action.ActionAttack] = 22
	attackFrames[1][action.ActionCharge] = 24
	attackFrames[1][action.ActionSkill] = 11
	attackFrames[1][action.ActionBurst] = 11
	attackFrames[1][action.ActionDash] = 10
	attackFrames[1][action.ActionJump] = 10
	attackFrames[1][action.ActionSwap] = 8

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2], 67) // walk
	attackFrames[2][action.ActionAttack] = 54
	attackFrames[2][action.ActionCharge] = 52
	attackFrames[2][action.ActionSkill] = 32
	attackFrames[2][action.ActionBurst] = 33
	attackFrames[2][action.ActionDash] = 32
	attackFrames[2][action.ActionJump] = 34
	attackFrames[2][action.ActionSwap] = 32
}

func (c *char) Attack(p map[string]int) (action.Info, error) {

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       fmt.Sprintf("Normal %v", c.NormalCounter),
		AttackTag:  attacks.AttackTagNormal,
		ICDTag:     attacks.ICDTagNormalAttack,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
		Mult:       attack[c.NormalCounter][c.TalentLvlAttack()],
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(
			c.Core.Combat.PrimaryTarget(),
			nil,
			0.7,
		),
		attackHitmarks[c.NormalCounter],
		attackHitmarks[c.NormalCounter],
	)

	defer c.AdvanceNormalIndex()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackFrames[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackFrames[c.NormalCounter][action.ActionSwap],
		State:           action.NormalAttackState,
	}, nil
}
