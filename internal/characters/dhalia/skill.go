package dhalia

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var (
	skillFrames []int
)

const (
	skillHitmark = 27
)

func init() {
	// Tap E
	skillFrames = frames.InitAbilSlice(32)
	skillFrames[action.ActionBurst] = 33
	skillFrames[action.ActionDash] = 33
}

func (c *char) Skill(p map[string]int) (action.Info, error) {

	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Flipclaw Strike",
		AttackTag:          attacks.AttackTagElementalArt,
		ICDTag:             attacks.ICDTagNone,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeDefault,
		Element:            attributes.Dendro,
		Durability:         25,
		Mult:               skill[c.TalentLvlSkill()],
		HitlagHaltFrames:   0.1 * 60,
		HitlagFactor:       0.01,
		CanBeDefenseHalted: true,
	}
	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0.5}, 2.6)
	c.Core.QueueAttack(
		ai,
		ap,
		skillHitmark,
		skillHitmark,
	)

	c.SetCDWithDelay(action.ActionSkill, 9*60, 1)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionSwap], // earliest cancel
		State:           action.SkillState,
	}, nil
}

