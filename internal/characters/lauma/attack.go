package lauma

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

const normalHitNum = 3

var (
	attackFrames   [][]int
	attackHitmarks = []int{15, 22, 13}
	attackHitboxes = [][]float64{{2, 3}, {2, 3}, {2, 3}}
	attackOffsets  = []float64{-0.2, -0.2, -0.2}
)

func init() {
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0], 18) // N1 -> D
	attackFrames[0][action.ActionAttack] = 27

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1], 27) // N2 -> D
	attackFrames[1][action.ActionAttack] = 49

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2], 23) // N3 -> D
	attackFrames[2][action.ActionAttack] = 30
}

// Normal Attack
// Performs up to 3 attacks that deal Dendro DMG
func (c *char) Attack(p map[string]int) (action.Info, error) {
	// C6: Check if normal attack should be converted to Lunar-Bloom DMG
	if c.c6NormalAttackConversion() {
		// C6 conversion already handled, just return basic action info
		defer c.AdvanceNormalIndex()
		return action.Info{
			Frames:          frames.NewAttackFunc(c.Character, attackFrames),
			AnimationLength: attackFrames[c.NormalCounter][action.InvalidAction],
			CanQueueAfter:   attackHitmarks[c.NormalCounter],
			State:           action.NormalAttackState,
		}, nil
	}

	ai := combat.AttackInfo{
		ActorIndex:   c.Index,
		Abil:         fmt.Sprintf("Normal %v", c.NormalCounter),
		AttackTag:    attacks.AttackTagNormal,
		ICDTag:       attacks.ICDTagNormalAttack,
		ICDGroup:     attacks.ICDGroupDefault,
		StrikeType:   attacks.StrikeTypeDefault,
		Element:      attributes.Dendro,
		Durability:   25,
		Mult:         attack[c.NormalCounter][c.TalentLvlAttack()],
		HitlagFactor: 0.01,
	}

	ap := combat.NewBoxHitOnTarget(
		c.Core.Combat.Player(),
		geometry.Point{Y: attackOffsets[c.NormalCounter]},
		attackHitboxes[c.NormalCounter][0],
		attackHitboxes[c.NormalCounter][1],
	)
	c.Core.QueueAttack(ai, ap, attackHitmarks[c.NormalCounter], attackHitmarks[c.NormalCounter])

	defer c.AdvanceNormalIndex()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackFrames[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[c.NormalCounter],
		State:           action.NormalAttackState,
	}, nil
}

