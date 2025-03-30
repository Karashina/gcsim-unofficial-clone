package varesa

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
	attackFrames     [][]int
	attackHitmarks   = []int{31, 7, 33}
	attackFramesFP   [][]int
	attackHitmarksFP = []int{22, 38, 34}
)

func init() {
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0], 41) // N1 -> N2
	attackFrames[0][action.ActionJump] = 36

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1], 33) // N2 -> N3
	attackFrames[1][action.ActionJump] = 15

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2], 44) // N3 -> N1
	attackFrames[2][action.ActionJump] = 54

	attackFramesFP = make([][]int, normalHitNum)

	attackFramesFP[0] = frames.InitNormalCancelSlice(attackHitmarks[0], 27) // N1 -> N2
	attackFramesFP[0][action.ActionJump] = 22

	attackFramesFP[1] = frames.InitNormalCancelSlice(attackHitmarks[1], 54) // N2 -> N3
	attackFramesFP[1][action.ActionJump] = 38

	attackFramesFP[2] = frames.InitNormalCancelSlice(attackHitmarks[2], 66) // N3 -> N1
	attackFramesFP[2][action.ActionJump] = 34
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(skillKey) {
		c.DeleteStatus(skillKey)
		c.fastCharge = true
		return c.ChargeAttack(p)
	}
	if c.StatusIsActive(fieryPassionKey) {
		return c.AttackFP(p)
	}

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       fmt.Sprintf("Normal %v", c.NormalCounter),
		AttackTag:  attacks.AttackTagNormal,
		ICDTag:     attacks.ICDTagNormalAttack,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       attack[c.NormalCounter][c.TalentLvlAttack()],
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(
			c.Core.Combat.PrimaryTarget(),
			nil,
			2,
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

//---------------------Fiery Passion--------------------------//

func (c *char) AttackFP(p map[string]int) (action.Info, error) {

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       fmt.Sprintf("Normal %v (Fiery Passion)", c.NormalCounter),
		AttackTag:  attacks.AttackTagNormal,
		ICDTag:     attacks.ICDTagNormalAttack,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       attackfp[c.NormalCounter][c.TalentLvlAttack()],
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(
			c.Core.Combat.PrimaryTarget(),
			nil,
			2,
		),
		attackHitmarksFP[c.NormalCounter],
		attackHitmarksFP[c.NormalCounter],
	)

	defer c.AdvanceNormalIndex()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFramesFP),
		AnimationLength: attackFramesFP[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackFramesFP[c.NormalCounter][action.ActionSwap],
		State:           action.NormalAttackState,
	}, nil
}
