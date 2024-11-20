package chasca

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
	attackFrames   [][]int
	attackHitmarks = [][]int{{16}, {9}, {15, 21}, {39, 42}}
	tapFrames      []int
	tapHitmark     = 32
)

const normalHitNum = 4

func init() {
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0][0], 26)
	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1][0], 21)
	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2][1], 37)
	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmarks[3][1], 76)

	tapFrames = make([]int, 1)

	tapFrames = frames.InitAbilSlice(40)
	tapFrames[action.ActionAttack] = 40
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	if c.nightsoulState.HasBlessing() {
		return c.fireTap(), nil
	}

	travel, ok := p["travel"]
	if !ok {
		travel = 10
	}

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       fmt.Sprintf("Normal %v (NA)", c.NormalCounter),
		AttackTag:  attacks.AttackTagNormal,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypePierce,
		Element:    attributes.Physical,
		Durability: 25,
	}

	for i, mult := range attack[c.NormalCounter] {
		ai.Mult = mult[c.TalentLvlAttack()]
		c.Core.QueueAttack(
			ai,
			combat.NewBoxHit(
				c.Core.Combat.Player(),
				c.Core.Combat.PrimaryTarget(),
				geometry.Point{Y: -0.5},
				0.1,
				1,
			),
			attackHitmarks[c.NormalCounter][i],
			attackHitmarks[c.NormalCounter][i]+travel,
		)
	}

	defer c.AdvanceNormalIndex()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackFrames[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[c.NormalCounter][len(attackHitmarks[c.NormalCounter])-1],
		State:           action.NormalAttackState,
	}, nil
}

func (c *char) fireTap() action.Info {

	c.NormalCounter = 0

	c.QueueCharTask(func() {
		ai := combat.AttackInfo{
			ActorIndex:     c.Index,
			Abil:           "Multitarget Fire Tap DMG (E)",
			AttackTag:      attacks.AttackTagNormal,
			AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
			ICDTag:         attacks.ICDTagNone,
			ICDGroup:       attacks.ICDGroupDefault,
			StrikeType:     attacks.StrikeTypeDefault,
			Mult:           tap[c.TalentLvlSkill()],
			Element:        attributes.Anemo,
			Durability:     25,
		}

		primaryEnemy, ok := c.Core.Combat.PrimaryTarget().(combat.Enemy)
		if !ok {
			return
		}

		c.Core.QueueAttack(
			ai,
			combat.NewSingleTargetHit(primaryEnemy.Key()),
			0,
			0,
		)

	}, tapHitmark)

	return action.Info{
		Frames:          frames.NewAbilFunc(tapFrames),
		AnimationLength: tapFrames[action.WalkState],
		CanQueueAfter:   tapFrames[action.ActionAttack],
		State:           action.NormalAttackState,
	}
}
