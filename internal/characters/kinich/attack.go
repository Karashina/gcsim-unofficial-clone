package kinich

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
	attackFramesE         []int
	attackHitmarks        = []int{24, 25, 51}
	attackReleaseE        = []int{15, 23}
	attackHitlagHaltFrame = []float64{0.10, 0.09, 0.12}
	attackPoiseDMG        = []float64{130, 120, 160}
	attackHitboxes        = []float64{2, 2, 2}
	attackOffsets         = []float64{1, 1, 1}
	attackFanAngles       = []float64{270, 360, 90}
)

const normalHitNum = 3

func init() {
	attackFrames = make([][]int, normalHitNum)

	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0], 38)

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1], 44)

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2], 94)
	attackFrames[2][action.ActionCharge] = 500 // illegal action

	attackFramesE = make([]int, normalHitNum)
	attackFramesE = frames.InitAbilSlice(38) // jump cancel
	attackFramesE[action.ActionDash] = 24
	attackFramesE[action.ActionAttack] = 40
	attackFramesE[action.ActionSkill] = 35
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(skillKey) {
		return c.AttackLoopShot(p)
	}
	ai := combat.AttackInfo{
		Abil:               fmt.Sprintf("Normal %v", c.NormalCounter),
		ActorIndex:         c.Index,
		AttackTag:          attacks.AttackTagNormal,
		ICDTag:             attacks.ICDTagNormalAttack,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeBlunt,
		PoiseDMG:           attackPoiseDMG[c.NormalCounter],
		Element:            attributes.Physical,
		Durability:         25,
		Mult:               attack[c.NormalCounter][c.TalentLvlAttack()],
		HitlagFactor:       0.01,
		HitlagHaltFrames:   attackHitlagHaltFrame[c.NormalCounter] * 60,
		CanBeDefenseHalted: true,
	}
	ap := combat.NewCircleHitOnTargetFanAngle(
		c.Core.Combat.Player(),
		geometry.Point{Y: attackOffsets[c.NormalCounter]},
		attackHitboxes[c.NormalCounter],
		attackFanAngles[c.NormalCounter],
	)
	if c.NormalCounter == 2 {
		ap = combat.NewBoxHitOnTarget(
			c.Core.Combat.Player(),
			geometry.Point{Y: attackOffsets[c.NormalCounter]},
			attackHitboxes[c.NormalCounter],
			attackHitboxes[c.NormalCounter],
		)
	}
	c.Core.QueueAttack(ai, ap, attackHitmarks[c.NormalCounter], attackHitmarks[c.NormalCounter])

	defer c.AdvanceNormalIndex()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackFrames[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[c.NormalCounter],
		State:           action.NormalAttackState,
	}, nil
}

func (c *char) AttackLoopShot(p map[string]int) (action.Info, error) {
	if p["blindspot"] > 0 && c.StatusIsActive(blindspotKey) {
		c.AddNightsoul("kinich-blindspot", 4)
		c.DeleteStatus(blindspotKey)
	}
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Canopy Hunter: Riding High (Loop Shot DMG)",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagKinichSKill,
		ICDGroup:   attacks.ICDGroupKinichSkill,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Dendro,
		Durability: 25,
		Mult:       skillNA[c.TalentLvlSkill()],
		Alignment:  attacks.AdditionalTagNightsoul,
	}
	ap := combat.NewSingleTargetHit(c.Core.Combat.PrimaryTarget().Key())

	c.AddNightsoul("kinich-loopshot", 3)
	c.Core.QueueAttack(ai, ap, attackReleaseE[0], attackReleaseE[0]+c.skilltravel, c.c2CB())
	c.Core.QueueAttack(ai, ap, attackReleaseE[1], attackReleaseE[1]+c.skilltravel, c.c2CB())

	c.AddStatus(skillLinkKey, 1*60, true)
	c.c4energy()

	return action.Info{
		Frames:          frames.NewAbilFunc(attackFramesE),
		AnimationLength: attackFramesE[action.InvalidAction],
		CanQueueAfter:   attackFramesE[action.ActionSwap],
		State:           action.NormalAttackState,
	}, nil
}
