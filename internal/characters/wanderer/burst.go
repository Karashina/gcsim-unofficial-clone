package wanderer

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
)

var burstFramesNormal []int
var burstFramesE []int

func init() {
	burstFramesNormal = frames.InitAbilSlice(101)
	burstFramesNormal[action.ActionAttack] = 94
	burstFramesNormal[action.ActionCharge] = 96
	burstFramesNormal[action.ActionSkill] = 95
	burstFramesNormal[action.ActionDash] = 97
	burstFramesNormal[action.ActionJump] = 96
	burstFramesNormal[action.ActionSwap] = 94

	// 交代時の落下を含む
	burstFramesE = frames.InitAbilSlice(145)
	burstFramesE[action.ActionAttack] = 117
	burstFramesE[action.ActionCharge] = 119
	burstFramesE[action.ActionDash] = 119
	burstFramesE[action.ActionJump] = 119
	burstFramesE[action.ActionWalk] = 117
}

// 最初のヒットマーク
const burstHitmark = 92

// 追加ヒット間の遅延
const burstHitmarkDelay = 6

// スナップショット段階までのフレーム数
// TODO: 正確なフレームを確認
const burstSnapshotDelay = 55

func (c *char) Burst(p map[string]int) (action.Info, error) {
	delay := c.checkForSkillEnd()

	if c.StatusIsActive(skillKey) {
		// delay == 0の場合のみ発生するため、無視可能
		return c.WindfavoredBurst(p)
	}

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Kyougen: Five Ceremonial Plays",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       burst[c.TalentLvlBurst()],
	}

	for i := 0; i < 5; i++ {
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 5),
			delay+burstSnapshotDelay, delay+burstHitmark+i*burstHitmarkDelay)
	}

	c.SetCD(action.ActionBurst, 15*60)
	c.ConsumeEnergy(5)

	return action.Info{
		Frames:          func(next action.Action) int { return delay + burstFramesNormal[next] },
		AnimationLength: delay + burstFramesNormal[action.InvalidAction],
		CanQueueAfter:   delay + burstFramesNormal[action.ActionAttack],
		State:           action.BurstState,
	}, nil
}

func (c *char) WindfavoredBurst(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Kyougen: Five Ceremonial Plays (Windfavored)",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       burst[c.TalentLvlBurst()],
	}

	c.c2()

	for i := 0; i < 5; i++ {
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 5),
			burstSnapshotDelay, burstHitmark+i*burstHitmarkDelay)
	}

	c.SetCD(action.ActionBurst, 15*60)
	c.ConsumeEnergy(5)

	// 必要（SwapStateへの遷移がそれ以外に不可能なため）
	c.Core.Player.SwapCD = 26
	// ここで空居ポイントをリセット
	c.skydwellerPoints = 0

	return action.Info{
		Frames:          func(next action.Action) int { return burstFramesE[next] },
		AnimationLength: burstFramesE[action.InvalidAction],
		CanQueueAfter:   burstFramesE[action.ActionWalk],
		State:           action.BurstState,
		OnRemoved: func(next action.AnimationState) {
			if next == action.SwapState {
				c.checkForSkillEnd()
			}
		},
	}, nil
}
