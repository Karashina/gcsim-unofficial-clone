package ayaka

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var burstFrames []int

const burstHitmark = 104

func init() {
	burstFrames = frames.InitAbilSlice(125) // Q -> D
	burstFrames[action.ActionAttack] = 124  // Q -> N1
	burstFrames[action.ActionSkill] = 124   // Q -> E
	burstFrames[action.ActionJump] = 113    // Q -> J
	burstFrames[action.ActionSwap] = 123    // Q -> Swap
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		Abil:       "Soumetsu",
		ActorIndex: c.Index,
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		Element:    attributes.Cryo,
		Durability: 25,
	}

	// 5秒間、20ティック、つまり15フレームごとに1回、5秒後にブルーム
	ai.Mult = burstBloom[c.TalentLvlBurst()]
	ai.StrikeType = attacks.StrikeTypeDefault
	ai.Abil = "Soumetsu (Bloom)"
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHit(
			c.Core.Combat.Player(),
			c.Core.Combat.PrimaryTarget(),
			nil,
			5,
		),
		burstHitmark,
		burstHitmark+300,
		c.c4,
	)

	// 2凸ミニ霜華のブルーム
	var aiC2 combat.AttackInfo
	if c.Base.Cons >= 2 {
		aiC2 = ai
		aiC2.Mult = burstBloom[c.TalentLvlBurst()] * .2
		aiC2.Abil = "C2 Mini-Frostflake Seki no To (Bloom)"
		// TODO: 位置/サイズが不明確...
		for i := 0; i < 2; i++ {
			c.Core.QueueAttack(
				aiC2,
				combat.NewCircleHit(
					c.Core.Combat.Player(),
					c.Core.Combat.PrimaryTarget(),
					nil,
					3,
				),
				burstHitmark,
				burstHitmark+300,
				c.c4,
			)
		}
	}

	for i := 0; i < 19; i++ {
		ai.Mult = burstCut[c.TalentLvlBurst()]
		ai.StrikeType = attacks.StrikeTypeSlash
		ai.Abil = "Soumetsu (Cutting)"
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHit(
				c.Core.Combat.Player(),
				c.Core.Combat.PrimaryTarget(),
				geometry.Point{Y: 0.3},
				3,
			),
			burstHitmark,
			burstHitmark+i*15,
			c.c4,
		)

		// 2凸ミニ霜華のカッティング
		if c.Base.Cons >= 2 {
			aiC2.Mult = burstCut[c.TalentLvlBurst()] * .20
			aiC2.StrikeType = attacks.StrikeTypeSlash
			aiC2.Abil = "C2 Mini-Frostflake Seki no To (Cutting)"
			// TODO: 位置/サイズが不明確...
			for j := 0; j < 2; j++ {
				c.Core.QueueAttack(
					aiC2,
					combat.NewCircleHit(
						c.Core.Combat.Player(),
						c.Core.Combat.PrimaryTarget(),
						geometry.Point{Y: 0.3},
						1.5,
					),
					burstHitmark,
					burstHitmark+i*15,
					c.c4,
				)
			}
		}
	}

	c.ConsumeEnergy(8)
	c.SetCD(action.ActionBurst, 20*60)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionJump], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}
