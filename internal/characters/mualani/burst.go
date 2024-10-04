package mualani

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/glog"
)

var burstFramesNormal []int

func init() {
	burstFramesNormal = frames.InitAbilSlice(174)
	burstFramesNormal[action.ActionAttack] = 174
	burstFramesNormal[action.ActionCharge] = 174
	burstFramesNormal[action.ActionSkill] = 174
	burstFramesNormal[action.ActionDash] = 174
	burstFramesNormal[action.ActionJump] = 174
	burstFramesNormal[action.ActionSwap] = 174
}

const burstHitmark = 116

func (c *char) Burst(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Boomsharka-laka",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
		FlatDmg:    c.MaxHP()*burst[c.TalentLvlBurst()] + c.a4buff,
		Alignment:  attacks.AdditionalTagNightsoul,
	}

	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 5),
		burstHitmark, burstHitmark)

	c.SetCD(action.ActionBurst, 15*60)
	c.Core.Log.NewEvent("burst a4 buff", glog.LogCharacterEvent, c.Index).
		Write("a4 count", c.a4stacks)
	c.ConsumeEnergy(7)

	return action.Info{
		Frames:          func(next action.Action) int { return burstFramesNormal[next] },
		AnimationLength: burstFramesNormal[action.InvalidAction],
		CanQueueAfter:   burstFramesNormal[action.ActionAttack],
		State:           action.BurstState,
	}, nil
}
