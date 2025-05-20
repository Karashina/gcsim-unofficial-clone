package escoffier

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

var burstFrames []int

const (
	burstHitmark     = 92
	burstEnergyFrame = 3
)

func init() {
	burstFrames = frames.InitAbilSlice(109)
	burstFrames[action.ActionDash] = 105
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Scoring Cuts (Q)",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeSlash,
		Element:    attributes.Cryo,
		Durability: 50,
		Mult:       burst[c.TalentLvlBurst()],
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player().Pos(), nil, 12),
		burstHitmark,
		burstHitmark,
	)

	c.QueueCharTask(func() {
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index,
			Target:  -1,
			Message: "Scoring Cuts (Q-Heal)",
			Src:     burstHealCst[c.TalentLvlBurst()] + c.TotalAtk()*burstHealPct[c.TalentLvlBurst()],
			Bonus:   c.Stat(attributes.Heal),
		})
	}, 94)

	a1dur := 9 * 60
	c.c4count = 0
	if c.Base.Cons >= 4 {
		a1dur = 15 * 60
	}
	c.AddStatus(a1Key, a1dur, true)
	c.a1Src = c.Core.F
	c.QueueCharTask(c.a1(c.a1Src), a1InitialHeal)

	c.ConsumeEnergy(burstEnergyFrame)
	c.SetCD(action.ActionBurst, 15*60)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap],
		State:           action.BurstState,
	}, nil
}
