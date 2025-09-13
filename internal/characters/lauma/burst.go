package lauma

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
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
	// Count pyro characters in team for scaling
	pyroCount := 0
	for _, char := range c.Core.Player.Chars() {
		if char.Base.Element == attributes.Pyro {
			pyroCount++
		}
	}
	
	// Base burst damage scales with pyro characters in team
	burstMult := burst[c.TalentLvlBurst()] * (1.0 + float64(pyroCount-1)*0.2) // 20% more per additional pyro character
	
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Elemental Burst",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Pyro,
		Durability: 50,
		Mult:       burstMult,
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 4}, 4.0),
		burstHitmark,
		burstHitmark,
	)

	// Create pyro infusion field for team for 8 seconds if A4 is unlocked
	if c.Base.Ascension >= 4 {
		c.AddStatus("lauma-burst-field", 8*60, true)
		
		// Apply pyro infusion to team while field is active
		c.Core.Events.Subscribe(event.OnAttack, func(args ...interface{}) bool {
			if !c.StatusIsActive("lauma-burst-field") {
				return false
			}
			
			atk := args[1].(*combat.AttackEvent)
			if atk.Info.Element == attributes.Physical {
				atk.Info.Element = attributes.Pyro
			}
			
			return false
		}, "lauma-burst-infusion")
	}

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