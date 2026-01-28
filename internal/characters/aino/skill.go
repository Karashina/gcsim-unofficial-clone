package aino

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var skillFrames []int

const (
	skillStage1Hitmark = 16
	skillStage2Hitmark = 36
)

func init() {
	skillFrames = frames.InitAbilSlice(46)
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
