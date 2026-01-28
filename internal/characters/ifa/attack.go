package ifa

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
	attackFrames   [][]int
	attackHitmarks = []int{13, 11, 32}

	attackSkillTapFrames []int
)

const normalHitNum = 3
const attackSkillTapHitmark = 51

func init() {
	attackFrames = make([][]int, normalHitNum)
	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0], 20) // N1 -> Walk
	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1], 27) // N2 -> Walk
	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2], 32) // N3 -> Walk

	attackFrames[0][action.ActionDash] = 15
	attackFrames[1][action.ActionDash] = 19
	attackFrames[2][action.ActionDash] = 43

	attackSkillTapFrames = frames.InitAbilSlice(39)
}

// Normal attack damage queue generator
// relatively standard with no major differences versus other bow characters
// Has "travel" parameter, used to set the number of frames that the arrow is in the air (default = 10)
func (c *char) Attack(p map[string]int) (action.Info, error) {
	if c.nightsoulState.HasBlessing() {
		return c.attackSkillTap(p), nil
	}

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
		combat.NewBoxHit(
			c.Core.Combat.Player(),
			c.Core.Combat.PrimaryTarget(),
			geometry.Point{Y: -0.5},
			0.1,
			1,
		),
		attackHitmarks[c.NormalCounter],
		attackHitmarks[c.NormalCounter]+travel,
	)
	defer c.AdvanceNormalIndex()

	return action.Info{
		Frames:          func(next action.Action) int { return frames.NewAttackFunc(c.Character, attackFrames)(next) },
		AnimationLength: attackFrames[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[c.NormalCounter],
		State:           action.NormalAttackState,
	}, nil
}

func (c *char) attackSkillTap(_ map[string]int) action.Info {
	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Tonicshot (Tap)(E)",
		AttackTag:      attacks.AttackTagNormal,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagNormalAttack,
		ICDGroup:       attacks.ICDGroupIfaShots,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Anemo,
		Durability:     25,
		Mult:           tonicshot[c.TalentLvlSkill()],
	}

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 2.0)
	c.QueueCharTask(func() {
		if !c.nightsoulState.HasBlessing() {
			return
		}
		c.Core.QueueAttack(
			ai,
			ap,
			0,
			0,
			c.particleCB,
		)
	}, attackSkillTapHitmark)

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackSkillTapFrames[action.InvalidAction],
		CanQueueAfter:   1,
		State:           action.NormalAttackState,
	}
}

func (c *char) attackSkillHold(p map[string]int) action.Info {
	hold, ok := p["hold"]
	if !ok {
		hold = 590
	}
	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Tonicshot (Hold)(E)",
		AttackTag:      attacks.AttackTagNormal,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagNormalAttack,
		ICDGroup:       attacks.ICDGroupIfaShots,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Anemo,
		Durability:     25,
		Mult:           tonicshot[c.TalentLvlSkill()],
	}

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 2.0)
	c.QueueCharTask(func() {
		if !c.nightsoulState.HasBlessing() {
			return
		}
		c.Core.QueueAttack(
			ai,
			ap,
			0,
			0,
			c.particleCB,
		)
	}, chargeTonicInitialHitmark)
	k := 0
	for i := chargeTonicInitialHitmark; i < chargeTonicInitialHitmark+hold; i = i + chargeTonicInterval {
		c.QueueCharTask(func() {
			if !c.nightsoulState.HasBlessing() {
				return
			}
			ai := combat.AttackInfo{
				ActorIndex:     c.Index,
				Abil:           "Tonicshot (Hold)(E)",
				AttackTag:      attacks.AttackTagNormal,
				AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
				ICDTag:         attacks.ICDTagNormalAttack,
				ICDGroup:       attacks.ICDGroupIfaShots,
				StrikeType:     attacks.StrikeTypeDefault,
				Element:        attributes.Anemo,
				Durability:     25,
				Mult:           tonicshot[c.TalentLvlSkill()],
			}
			ap := combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 2.0)

			c.Core.QueueAttack(
				ai,
				ap,
				0,
				0,
				c.particleCB,
			)
			c.c6()
		}, chargeTonicInterval+k*chargeTonicInterval)
		k++
	}

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: min(chargemaxDur, attackSkillTapFrames[action.InvalidAction]+hold),
		CanQueueAfter:   chargeTonicInitialHitmark + hold,
		State:           action.NormalAttackState,
	}
}
