package flins

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
	attackFrames           [][]int
	attackHitmarks         = [][]int{{12}, {8}, {15}, {21, 31}, {31}}
	attackHitlagHaltFrame  = [][]float64{{0.12}, {0.06}, {0.03}, {0.00, 0.00}, {0.06}}
	attackFramesE          [][]int
	attackHitmarksE        = [][]int{{12}, {9}, {16}, {20, 31}, {31}}
	attackHitlagHaltFrameE = [][]float64{{0.12}, {0.00}, {0.00}, {0.00, 0.00}, {0.12}}
	attackDefHalt          = [][]bool{{true}, {true}, {true}, {false, true}, {true}}
	attackHitboxes         = [][][]float64{{{2}}, {{2}}, {{2}}, {{2.5}, {2.5}}, {{2.5}}}
	attackOffsets          = [][]float64{{-0.2}, {-0.2}, {-0.2}, {-0.2, -0.2}, {-0.2}}
)

const normalHitNum = 5

func init() {
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0][0], 21)
	attackFrames[0][action.ActionCharge] = 21
	attackFrames[0][action.ActionDash] = 14

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1][0], 18)
	attackFrames[1][action.ActionDash] = 19

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2][0], 26)
	attackFrames[2][action.ActionDash] = 22

	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmarks[3][1], 42)
	attackFrames[3][action.ActionDash] = 41

	attackFrames[4] = frames.InitNormalCancelSlice(attackHitmarks[4][0], 63)
	attackFrames[4][action.ActionDash] = 31
	attackFrames[4][action.ActionCharge] = 500 // Illegal action; needs better handling

	attackFramesE = make([][]int, normalHitNum)

	attackFramesE[0] = frames.InitNormalCancelSlice(attackHitmarks[0][0], 18)
	attackFramesE[0][action.ActionCharge] = 21
	attackFramesE[0][action.ActionDash] = 14

	attackFramesE[1] = frames.InitNormalCancelSlice(attackHitmarks[1][0], 20)
	attackFramesE[1][action.ActionDash] = 19

	attackFramesE[2] = frames.InitNormalCancelSlice(attackHitmarks[2][0], 24)
	attackFramesE[2][action.ActionDash] = 22

	attackFramesE[3] = frames.InitNormalCancelSlice(attackHitmarks[3][1], 40)
	attackFramesE[3][action.ActionDash] = 41

	attackFramesE[4] = frames.InitNormalCancelSlice(attackHitmarks[4][0], 65)
	attackFramesE[4][action.ActionDash] = 37
	attackFramesE[4][action.ActionCharge] = 500 // Illegal action; needs better handling
}

// Normal attack implementation
// Pocztowy Demonspear
func (c *char) Attack(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(skillKey) {
		return c.attackE() // go to Electro-infused attacks during Manifest Flame form
	}
	for i, mult := range attack[c.NormalCounter] {
		ai := combat.AttackInfo{
			ActorIndex:         c.Index,
			Abil:               fmt.Sprintf("Pocztowy Demonspear %v", c.NormalCounter),
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

		ap := combat.NewCircleHitOnTarget(
			c.Core.Combat.Player(),
			geometry.Point{Y: attackOffsets[c.NormalCounter][i]},
			attackHitboxes[c.NormalCounter][i][0],
		)
		if c.NormalCounter == 0 {
			ai.StrikeType = attacks.StrikeTypeSpear
			// Check length before accessing [1]
			hitbox := attackHitboxes[c.NormalCounter][i]
			var width float64 = 1.0
			if len(hitbox) > 1 {
				width = hitbox[1]
			}
			ap = combat.NewBoxHitOnTarget(
				c.Core.Combat.Player(),
				geometry.Point{Y: attackOffsets[c.NormalCounter][i]},
				hitbox[0],
				width,
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

// Normal attack implementation during Manifest Flame form (Electro-infused)
// Ancient Rite: Arcane Light
func (c *char) attackE() (action.Info, error) {
	c2CB := c.c2AdditionalDamage()
	for i, mult := range attack_e[c.NormalCounter] {
		ai := combat.AttackInfo{
			ActorIndex:         c.Index,
			Abil:               fmt.Sprintf("Ancient Rite: Arcane Light %v", c.NormalCounter),
			AttackTag:          attacks.AttackTagNormal,
			ICDTag:             attacks.ICDTagNormalAttack,
			ICDGroup:           attacks.ICDGroupDefault,
			StrikeType:         attacks.StrikeTypeSlash,
			Element:            attributes.Electro,
			Durability:         25,
			Mult:               mult[c.TalentLvlSkill()],
			HitlagFactor:       0.01,
			HitlagHaltFrames:   attackHitlagHaltFrameE[c.NormalCounter][i] * 60,
			CanBeDefenseHalted: attackDefHalt[c.NormalCounter][i],
			IgnoreInfusion:     true,
		}

		ap := combat.NewCircleHitOnTarget(
			c.Core.Combat.Player(),
			geometry.Point{Y: attackOffsets[c.NormalCounter][i]},
			attackHitboxes[c.NormalCounter][i][0],
		)
		if c.NormalCounter == 0 {
			ai.StrikeType = attacks.StrikeTypeSpear
			// Check length before accessing [1]
			hitbox := attackHitboxes[c.NormalCounter][i]
			var width float64 = 1.0
			if len(hitbox) > 1 {
				width = hitbox[1]
			}
			ap = combat.NewBoxHitOnTarget(
				c.Core.Combat.Player(),
				geometry.Point{Y: attackOffsets[c.NormalCounter][i]},
				hitbox[0],
				width,
			)
		}
		c.QueueCharTask(func() {
			c.Core.QueueAttack(ai, ap, 0, 0, c.particleCB, c2CB)
		}, attackHitmarksE[c.NormalCounter][i])
	}

	defer c.AdvanceNormalIndex()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFramesE),
		AnimationLength: attackFramesE[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackHitmarksE[c.NormalCounter][len(attackHitmarksE[c.NormalCounter])-1],
		State:           action.NormalAttackState,
	}, nil
}
