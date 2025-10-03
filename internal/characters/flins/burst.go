package flins

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
)

var burstFrames []int

const (
	hitmark = 119
)

func init() {
	burstFrames = frames.InitAbilSlice(114)
}

// Burst ability implementation
// Ancient Ritual: Cometh the Night (Q)
// Flins unleashes the true power of his lamp, dealing AoE Electro DMG
// Transforms to Thunderous Symphony after using Northland Spearstorm
func (c *char) Burst(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Ancient Ritual: Cometh the Night (Q)",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       burst[c.TalentLvlBurst()],
	}
	c.QueueCharTask(func() {
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: -1.5}, 7),
			0, 0,
		)
	}, hitmark)

	c.a4()
	c.c2()
	c.SetCD(action.ActionBurst, 15*60)

	// resummon E
	c.skillSrc = c.Core.F
	for i := 0.0; i < skillTicks; i++ {
		c.Core.Tasks.Add(c.skillTick(c.skillSrc), skillFirstTickDelay+ceil(skillInterval*i))
	}
	c.AddStatus(skillKey, skillFirstTickDelay+ceil((skillTicks-1)*skillInterval), false)

	c.ConsumeEnergy(4)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // earliest cancel
		State:           action.BurstState,
	}, nil
}
