package aino

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
)

var skillFrames []int

const (
	skillStage1Hitmark = 30
	skillStage2Hitmark = 45
)

func init() {
	skillFrames = frames.InitAbilSlice(70)
	skillFrames[action.ActionAttack] = 60
	skillFrames[action.ActionBurst] = 60
	skillFrames[action.ActionDash] = 50
	skillFrames[action.ActionSwap] = 58
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	// Stage 1: Single target damage
	aiStage1 := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Musecatcher (Stage 1)",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
		Mult:       skillStage1[c.TalentLvlSkill()],
	}
	c.Core.QueueAttack(
		aiStage1,
		combat.NewSingleTargetHit(c.Core.Combat.PrimaryTarget().Key()),
		skillStage1Hitmark,
		skillStage1Hitmark,
	)

	// Stage 2: AoE damage
	aiStage2 := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Musecatcher (Stage 2)",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagElementalArt,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
		Mult:       skillStage2[c.TalentLvlSkill()],
	}
	c.Core.QueueAttack(
		aiStage2,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, 5),
		skillStage2Hitmark,
		skillStage2Hitmark,
	)

	c.SetCD(action.ActionSkill, 10*60)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash],
		State:           action.SkillState,
	}, nil
}
