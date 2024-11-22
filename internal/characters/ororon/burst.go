package ororon

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
)

const (
	burstInitialHitmarks    = 41
	burstDoTInitialHitmarks = 63
	burstDoTInterval        = 59
	burstDoTIntervalC4      = 39
)

var (
	burstFrames []int
)

func init() {
	burstFrames = frames.InitAbilSlice(180) // charge
	burstFrames[action.ActionAttack] = 167
	burstFrames[action.ActionSkill] = 166
	burstFrames[action.ActionDash] = 167
	burstFrames[action.ActionJump] = 167
	burstFrames[action.ActionWalk] = 167
	burstFrames[action.ActionSwap] = 108
}

func (c *char) Burst(p map[string]int) (action.Info, error) {

	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Dark Voices Echo: Ritual DMG (Q)",
		AttackTag:      attacks.AttackTagElementalBurst,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagNone,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Electro,
		Durability:     25,
		Mult:           burst[c.TalentLvlBurst()],
	}
	burstArea := combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 7)
	c.Core.QueueAttack(ai, burstArea, burstInitialHitmarks, burstInitialHitmarks)

	aidot := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Dark Voices Echo: Soundwave Collision DMG (Q)",
		AttackTag:      attacks.AttackTagElementalBurst,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagElementalBurst,
		ICDGroup:       attacks.ICDGroupDoriBurst,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Electro,
		Durability:     25,
		Mult:           burst[c.TalentLvlBurst()],
	}

	bursthitcount := 9
	interval := burstDoTInterval
	if c.Base.Cons >= 4 {
		bursthitcount = 11
		interval = burstDoTIntervalC4
	}
	for i := 0; i < bursthitcount; i++ {
		hitmark := burstDoTInitialHitmarks + interval*i
		c.QueueCharTask(func() {
			c.Core.QueueAttack(aidot, burstArea, 0, 0)
		}, hitmark)
	}

	c.c2Init()
	c.c6()
	c.SetCDWithDelay(action.ActionBurst, 15*60, 0)
	c.ConsumeEnergy(7)

	c.QueueCharTask(c.c4Energy, 9)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // earliest cancel
		State:           action.BurstState,
	}, nil
}

func (c *char) c4Energy() {
	if c.Base.Cons >= 4 {
		c.AddEnergy("ororon-c4", 8)
	}
}
