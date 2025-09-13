package lauma

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
)

var (
	attackFrames   [][]int
	attackHitmarks = []int{15, 20, 25, 35}
	normalHitNum   = len(attackHitmarks)
)

const normalHitICDKey = "lauma-normal"

func init() {
	attackFrames = make([][]int, normalHitNum)

	// Attack 1
	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0], 35)
	// Attack 2  
	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1], 40)
	// Attack 3
	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2], 45)
	// Attack 4
	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmarks[3], 55)
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Normal Attack",
		AttackTag:  attacks.AttackTagNormal,
		ICDTag:     attacks.ICDTagNormalAttack,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Physical,
		Durability: 25,
		Mult:       attack[c.NormalCounter][c.TalentLvlAttack()],
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, 0.8),
		attackHitmarks[c.NormalCounter],
		attackHitmarks[c.NormalCounter],
	)

	defer c.AdvanceNormalIndex()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackFrames[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[c.NormalCounter],
		State:           action.NormalAttackState,
	}, nil
}

// attack[NormalCounter][TalentLevel-1]
var attack = [][]float64{
	{0.4712, 0.5096, 0.5480, 0.5890, 0.6274, 0.6658, 0.7068, 0.7478, 0.7888, 0.8299, 0.8709},
	{0.4536, 0.4906, 0.5276, 0.5670, 0.6040, 0.6410, 0.6804, 0.7198, 0.7592, 0.7986, 0.8380},
	{0.5564, 0.6020, 0.6476, 0.6956, 0.7412, 0.7868, 0.8348, 0.8828, 0.9308, 0.9788, 1.0268},
	{0.7084, 0.7665, 0.8246, 0.8854, 0.9435, 1.0016, 1.0624, 1.1232, 1.1840, 1.2448, 1.3056},
}