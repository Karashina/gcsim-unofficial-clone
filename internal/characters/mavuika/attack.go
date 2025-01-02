package mavuika

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
	attackHitmarks        = [][]int{{22}, {13, 28}, {29, 34, 40}, {27}}
	attackPoiseDMG        = []float64{107, 48.8, 44.43, 155.4}
	attackHitlagHaltFrame = [][]float64{{0.09}, {0.05, 0.05}, {0.02, 0.02, 0.02}, {0.1}}
	attackDefHalt         = [][]bool{{true}, {true, false}, {true, false, false}, {true}}
	attackHitboxes        = []float64{2, 2.5, 3, 3}
	attackOffsets         = []float64{0.5, 0.5, 0.5, 0.5}
	attackEarliestCancel  = []int{22, 13, 29, 27}

	bikeAttackFrames   [][]int
	bikeAttackHitmarks = []int{21, 24, 32, 16, 40}
	bikeattackHitboxes = []float64{3, 3, 3, 3, 3}
	bikeattackOffsets  = []float64{0.5, 0.5, 0.5, 0.5, 0.5}
)

const (
	normalHitNum = 4
	skillHitNum  = 5
)

func init() {
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackEarliestCancel[0], 39)
	attackFrames[1] = frames.InitNormalCancelSlice(attackEarliestCancel[1], 46)
	attackFrames[2] = frames.InitNormalCancelSlice(attackEarliestCancel[2], 51)
	attackFrames[3] = frames.InitNormalCancelSlice(attackEarliestCancel[3], 61)

	// Skill attack
	bikeAttackFrames = make([][]int, skillHitNum)

	bikeAttackFrames[0] = frames.InitNormalCancelSlice(bikeAttackHitmarks[0], 30)
	bikeAttackFrames[1] = frames.InitNormalCancelSlice(bikeAttackHitmarks[1], 37)
	bikeAttackFrames[2] = frames.InitNormalCancelSlice(bikeAttackHitmarks[2], 37)
	bikeAttackFrames[3] = frames.InitNormalCancelSlice(bikeAttackHitmarks[3], 30)
	bikeAttackFrames[4] = frames.InitNormalCancelSlice(bikeAttackHitmarks[4], 68)
	bikeAttackFrames[4][action.ActionDash] = 48
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(bikeKey) {
		return c.skillAttack(p)
	}
	for i, mult := range attack[c.NormalCounter] {
		ai := combat.AttackInfo{
			ActorIndex:         c.Index,
			Abil:               fmt.Sprintf("Normal %v", c.NormalCounter),
			AttackTag:          attacks.AttackTagNormal,
			ICDTag:             attacks.ICDTagNormalAttack,
			ICDGroup:           attacks.ICDGroupDefault,
			StrikeType:         attacks.StrikeTypeBlunt,
			PoiseDMG:           attackPoiseDMG[c.NormalCounter],
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
		c.Core.QueueAttack(ai, ap, attackEarliestCancel[c.NormalCounter], attackHitmarks[c.NormalCounter][i])
	}

	defer c.AdvanceNormalIndex()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackFrames[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackEarliestCancel[c.NormalCounter],
		State:           action.NormalAttackState,
	}, nil
}

func (c *char) skillAttack(_ map[string]int) (action.Info, error) {
	buff := 0.0
	if c.StatusIsActive(BurstKey) {
		buff = c.consumedspirit
	}
	c2buff := 0.0
	if c.Base.Cons >= 2 {
		c2buff = 0.6
	}
	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           fmt.Sprintf("Flamestrider %d", c.normalBikeCounter),
		AttackTag:      attacks.AttackTagNormal,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagNormalAttack,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Pyro,
		Durability:     25,
		Mult:           bikeattack[c.normalBikeCounter][c.TalentLvlSkill()] + burstnabonus[c.TalentLvlBurst()]*buff + c2buff,
		IgnoreInfusion: true,
	}

	ap := combat.NewCircleHitOnTarget(
		c.Core.Combat.Player(),
		geometry.Point{Y: bikeattackOffsets[c.normalBikeCounter]},
		bikeattackHitboxes[c.normalBikeCounter],
	)

	c.QueueCharTask(func() {
		c.Core.QueueAttack(ai, ap, 0, 0)
	}, bikeAttackHitmarks[c.normalBikeCounter])

	c.reduceNightsoulPoints(1)

	defer c.AdvanceNormalIndex()

	return action.Info{
		Frames:          frames.NewAbilFunc(bikeAttackFrames[c.normalBikeCounter]),
		AnimationLength: bikeAttackFrames[c.normalBikeCounter][action.InvalidAction],
		CanQueueAfter:   bikeAttackFrames[c.normalBikeCounter][action.ActionBurst],
		State:           action.NormalAttackState,
	}, nil
}
