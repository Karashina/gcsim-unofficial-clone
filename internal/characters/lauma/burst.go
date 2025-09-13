package lauma

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
)

var burstFrames []int

const burstHitmark = 45

func init() {
	burstFrames = frames.InitAbilSlice(90) // total duration
	burstFrames[action.ActionAttack] = 50
	burstFrames[action.ActionSkill] = 50
	burstFrames[action.ActionDash] = burstHitmark
	burstFrames[action.ActionJump] = burstHitmark
	burstFrames[action.ActionSwap] = 75
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Elemental Burst",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Pyro,
		Durability: 50,
		Mult:       burst[c.TalentLvlBurst()],
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 4}, 4.0),
		burstHitmark,
		burstHitmark,
	)

	// Consume energy
	c.ConsumeEnergy(5)
	// Set burst cooldown  
	c.SetCDWithDelay(action.ActionBurst, 15*60, 45) // 15 second cooldown

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash],
		State:           action.BurstState,
	}, nil
}

// burst[TalentLevel-1]
var burst = []float64{3.776, 4.0592, 4.3424, 4.72, 5.0032, 5.2864, 5.664, 6.0416, 6.4192, 6.7968, 7.1744}