package citlali

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
)

const (
	burstHitmark = 118
	skullHitmark = 225
)

var (
	burstFrames []int
)

func init() {
	burstFrames = frames.InitAbilSlice(121)
}

func (c *char) Burst(p map[string]int) (action.Info, error) {

	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Edict of Entwined Splendor: Ice Storm DMG",
		AttackTag:      attacks.AttackTagElementalBurst,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagNone,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Cryo,
		Durability:     50,
		Mult:           burst[c.TalentLvlBurst()],
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 6.5), 0, burstHitmark)

	aiskull := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Edict of Entwined Splendor: Spiritvessel Skull DMG",
		AttackTag:      attacks.AttackTagElementalBurst,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagElementalBurst,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Cryo,
		Durability:     25,
		Mult:           burstskull[c.TalentLvlBurst()],
	}
	enemies := c.Core.Combat.RandomEnemiesWithinArea(combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 6.5), nil, 3)
	for _, enemy := range enemies {
		c.Core.QueueAttack(aiskull, combat.NewCircleHitOnTarget(enemy.Pos(), nil, 3.5), skullHitmark, skullHitmark)
	}
	c.SetCDWithDelay(action.ActionBurst, 15*60, 2)
	c.QueueCharTask(func() {
		if c.nightsoulState.HasBlessing() {
			c.nightsoulState.GeneratePoints(24)
		}
	}, 1)
	c.ConsumeEnergy(109)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // earliest cancel
		State:           action.BurstState,
	}, nil
}
