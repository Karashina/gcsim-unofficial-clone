package nefer

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
)

var (
	skillFrames []int
)

const (
	skillHitmark = 28
	skillKey     = "nefer-skill"
)

func init() {
	skillFrames = frames.InitAbilSlice(30)
}

// Elemental Skill
func (c *char) Skill(p map[string]int) (action.Info, error) {
	/*
		Nefer deals AoE Dendro DMG and entering the Shadow Dance state.
		Two initial charges.

		While in the Shadow Dance state, if you have at least 1 Verdant Dew, Nefer's Charged Attacks will be replaced with the special Charged Attack Phantasm Performance, which will not consume Stamina.
		TO COPILOT: this ability must be implemented in the charge.go but get the talent level for the Phantasm Performance from c.TalentLvlSkill().
	*/

	c.SetCD(action.ActionSkill, 9*60)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionSwap],
		State:           action.SkillState,
	}, nil
}
