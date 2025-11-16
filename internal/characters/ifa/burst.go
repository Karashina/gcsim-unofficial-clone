package ifa

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
)

var burstFramesGrounded []int

const (
	burstHitmark       = 43
	burstHitmarkSecond = 80
)

func init() {
	burstFramesGrounded = frames.InitAbilSlice(85)
	burstFramesGrounded[action.ActionDash] = 85 // Q -> D
}

func (c *char) Burst(p map[string]int) (action.Info, error) {

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 6.0)
	enemies := c.Core.Combat.EnemiesWithinArea(ap, nil)
	checkele := make([]attributes.Element, 4)

	j := 0
	for _, enemy := range enemies {
		if enemy == nil {
			break
		}
		checkele[j] = c.Core.Combat.AbsorbCheck(c.Index, combat.NewCircleHitOnTarget(enemy.Pos(), nil, 0.5), attributes.Pyro, attributes.Hydro, attributes.Electro, attributes.Cryo)
		j++
		if j >= 4 {
			break
		}
	}

	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Galesplitting Soulseeker Shell(Q)",
		AttackTag:      attacks.AttackTagElementalBurst,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagNone,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Anemo,
		Durability:     25,
		Mult:           burst[c.TalentLvlBurst()],
	}

	c.Core.QueueAttack(
		ai,
		ap,
		burstHitmark,
		burstHitmark,
	)

	for i, ele := range checkele {
		if ele == attributes.NoElement {
			continue
		}
		if i >= len(enemies) || enemies[i] == nil {
			continue
		}
		ai = combat.AttackInfo{
			ActorIndex:     c.Index,
			Abil:           "Sedation Mark(Q)",
			AttackTag:      attacks.AttackTagElementalBurst,
			AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
			ICDTag:         attacks.ICDTagNone,
			ICDGroup:       attacks.ICDGroupDefault,
			StrikeType:     attacks.StrikeTypeDefault,
			Element:        ele,
			Durability:     25,
			Mult:           sedationmark[c.TalentLvlBurst()],
		}
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(enemies[i].Pos(), nil, 0.5),
			burstHitmarkSecond,
			burstHitmarkSecond,
		)
	}

	c.c4()
	c.SetCD(action.ActionBurst, 15*60)
	c.ConsumeEnergy(6)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFramesGrounded),
		AnimationLength: burstFramesGrounded[action.InvalidAction],
		CanQueueAfter:   burstFramesGrounded[action.ActionDash], // earliest cancel
		State:           action.BurstState,
	}, nil
}

