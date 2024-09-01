package kachina

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
)

var burstFramesNormal []int

const (
	burstKey = "kachina-burst"
)

func init() {
	burstFramesNormal = frames.InitAbilSlice(67)
}

const burstHitmark = 40

func (c *char) Burst(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Time to Get Serious!",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		PoiseDMG:   150,
		Element:    attributes.Geo,
		Durability: 25,
		Mult:       burst[c.TalentLvlBurst()],
		UseDef:     true,
	}

	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 10),
		burstHitmark, burstHitmark)

	c.AddStatus(burstKey, 12*60, true)
	c.SetCD(action.ActionBurst, 18*60)
	c.c2()
	c.ConsumeEnergy(6)

	return action.Info{
		Frames:          func(next action.Action) int { return burstFramesNormal[next] },
		AnimationLength: burstFramesNormal[action.InvalidAction],
		CanQueueAfter:   burstFramesNormal[action.ActionAttack],
		State:           action.BurstState,
	}, nil
}
