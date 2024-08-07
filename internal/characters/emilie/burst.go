package emilie

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
	burstHitmark     = 1
	burstLCSpawn     = 316
	burstLCFirstTick = 89
	burstKey         = "emilie-burst"
)

func init() {
	burstFrames = frames.InitAbilSlice(103)
	burstFrames[action.ActionDash] = 103
	burstFrames[action.ActionJump] = 103
	burstFrames[action.ActionSwap] = 103
}

func (c *char) Burst(p map[string]int) (action.Info, error) {

	duration := 168
	targetinterval := 42
	if c.Base.Cons >= 4 {
		duration += 120
		targetinterval -= 18
	}
	// initial damage; part of the burst tag
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Level 3 Lumidouce Case Attack DMG (Q)",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypePierce,
		Element:    attributes.Dendro,
		Durability: 25,
		Mult:       burst[c.TalentLvlBurst()],
	}

	c.Core.Status.Add(burstKey, duration+127)
	if c.Base.Cons >= 6 {
		c.c6init()
	}

	burstArea := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 10)
	for i := 0; i < 2.8*60; i += 18 {
		c.Core.Tasks.Add(func() {
			// burst tick
			enemy := c.Core.Combat.RandomEnemyWithinArea(
				burstArea,
				func(e combat.Enemy) bool {
					return !e.StatusIsActive("emilie-burst-mark")
				},
			)
			var pos geometry.Point
			if enemy != nil {
				pos = enemy.Pos()
				enemy.AddStatus("emilie-burst-mark", 0.7*60, true) // same enemy can't be targeted again for 0.7s
			} else {
				pos = geometry.CalcRandomPointFromCenter(burstArea.Shape.Pos(), 0.5, 9.5, c.Core.Rand)
			}
			c.Core.QueueAttack(
				ai,
				combat.NewCircleHitOnTarget(pos, nil, 2.5),
				burstHitmark,
				burstHitmark,
			)
		}, i+127)
	}

	c.ConsumeEnergy(12)
	c.SetCD(action.ActionBurst, 13.5*60)

	c.LCActive = true
	c.burstLCSpawnSrc = c.Core.F
	burstLCFunc := c.burstLCSpawn(c.Core.F, 0, burstLCFirstTick)

	c.Core.Tasks.Add(burstLCFunc, burstLCSpawn)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // earliest cancel
		State:           action.BurstState,
		OnRemoved:       func(next action.AnimationState) { burstLCFunc() },
	}, nil
}

func (c *char) burstLCSpawn(src, LCSpawn, firstTick int) func() {
	return func() {
		if src != c.burstLCSpawnSrc {
			return
		}
		c.burstLCSpawnSrc = -1
		c.queueLumidouceCase("Burst", LCSpawn, firstTick)
	}
}
