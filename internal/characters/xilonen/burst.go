package xilonen

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

var burstFrames []int

const (
	bursthitmark     = 98
	burstDoThitmark1 = 137
	burstDoThitmark2 = 177
)

func init() {
	burstFrames = frames.InitAbilSlice(90) // Q -> D/J
}

func (c *char) Burst(p map[string]int) (action.Info, error) {

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Ocelotlicue Point!: Skill DMG",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		Element:    attributes.Geo,
		Durability: 25,
		Mult:       burst[c.TalentLvlBurst()],
		UseDef:     true,
		Alignment:  attacks.AdditionalTagNightsoul,
	}

	// initial hit
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 6), bursthitmark, bursthitmark)

	switch c.A1Mode {
	case 1:
		// DoT
		ai.Abil = "Ocelotlicue Point!: Follow-Up Beat DMG"
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 6), burstDoThitmark1, burstDoThitmark1)
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 6), burstDoThitmark2, burstDoThitmark2)
	case 2:
		for i := 0; i <= 8; i++ {
			c.Core.Tasks.Add(func() {
				c.Core.Player.Heal(info.HealInfo{
					Caller:  c.Index,
					Target:  c.Core.Player.ActiveChar().Index,
					Message: "Ocelotlicue Point!: Continuous Healing",
					Src:     burstheal[c.TalentLvlBurst()]*c.TotalDef() + bursthealconst[c.TalentLvlBurst()],
					Bonus:   c.Stat(attributes.Heal),
				})
			}, 246+89*i)
		}
	default:
		c.Core.Log.NewEvent("Invalid A1 Mode!", glog.LogCharacterEvent, c.Index)
	}

	c.ConsumeEnergy(10)
	c.SetCDWithDelay(action.ActionBurst, 15*60, 2)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // earliest cancel
		State:           action.BurstState,
	}, nil
}
