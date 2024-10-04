package kinich

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
)

var burstFramesNormal []int

func init() {
	burstFramesNormal = frames.InitAbilSlice(106)
}

const burstHitmark = 158
const dotHitmark = 240
const interval = 147

func (c *char) Burst(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Hail to the Almighty Dragonlord (Skill DMG)",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Dendro,
		Durability: 25,
		Mult:       burst[c.TalentLvlBurst()],
		Alignment:  attacks.AdditionalTagNightsoul,
	}
	aidot := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Hail to the Almighty Dragonlord (Dragon Breath DMG)",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Dendro,
		Durability: 25,
		Mult:       burstDot[c.TalentLvlBurst()],
		Alignment:  attacks.AdditionalTagNightsoul,
	}

	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 5),
		burstHitmark, burstHitmark)

	for i := dotHitmark; i < 6*interval+dotHitmark; i += interval {
		c.Core.QueueAttack(
			aidot,
			combat.NewBoxHit(c.Core.Combat.Player(), c.Core.Combat.PrimaryTarget(), nil, 2, 7),
			i,
			i,
		)
	}

	c.SetCD(action.ActionBurst, 17*60+2) // it seems his cd start from 17s????
	c.ExtendStatus(skillKey, 1.7*60)
	c.ConsumeEnergy(4)

	return action.Info{
		Frames:          func(next action.Action) int { return burstFramesNormal[next] },
		AnimationLength: burstFramesNormal[action.InvalidAction],
		CanQueueAfter:   burstFramesNormal[action.ActionAttack],
		State:           action.BurstState,
	}, nil
}
