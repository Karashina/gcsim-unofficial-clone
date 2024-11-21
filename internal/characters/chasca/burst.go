package chasca

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
)

var (
	burstFrames         []int
	burstInitialHitmark = 95
	burstHitmarks       = []int{134, 158, 175, 180, 183, 191}
)

func init() {
	burstFrames = frames.InitAbilSlice(100)
}

func (c *char) Burst(p map[string]int) (action.Info, error) {

	c.DeleteStatus(c4ICDKey)

	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Galesplitting Soulseeker Shell DMG (Q)",
		AttackTag:      attacks.AttackTagElementalBurst,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagNone,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Anemo,
		Durability:     25,
		Mult:           burst[c.TalentLvlBurst()],
	}
	burstArea := combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 6)

	c.Core.QueueAttack(ai, burstArea, burstInitialHitmark, burstInitialHitmark)

	elemcount := len(c.ElementSlot)
	elementUsage := make(map[attributes.Element]int)

	for i := 0; i < 6; i++ {
		if elemcount > 0 {
			ai.Abil = "Radiant Soulseeker Shell DMG (Q)"
			ai.ICDTag = attacks.ICDTagElementalBurst
			ai.ICDGroup = attacks.ICDGroupChascaConvertedShell
			ai.Mult = radiantsoulseeker[c.TalentLvlBurst()]

			var selectedElement attributes.Element
			found := false
			for attempts := 0; attempts < elemcount*10; attempts++ {
				selectedElement = c.ElementSlot[c.Core.Rand.Intn(len(c.ElementSlot))]
				if elementUsage[selectedElement] < 2 {
					found = true
					break
				}
			}
			if !found {
				elemcount = 0
				break
			}
			ai.Element = selectedElement
			elementUsage[selectedElement]++
		} else {
			ai.Abil = "Soulseeker Shell DMG (Q)"
			ai.ICDTag = attacks.ICDTagElementalBurst
			ai.ICDGroup = attacks.ICDGroupChascaConvertedShell
			ai.Mult = soulseeker[c.TalentLvlBurst()]
		}
		c.Core.QueueAttack(ai, combat.NewSingleTargetHit(c.Core.Combat.PrimaryTarget().Key()), burstHitmarks[i], burstHitmarks[i], c.c4CB)
	}

	c.SetCDWithDelay(action.ActionBurst, 15*60, 0)
	c.ConsumeEnergy(4)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // earliest cancel
		State:           action.BurstState,
	}, nil
}
