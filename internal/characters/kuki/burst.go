package kuki

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
)

var burstFrames []int

const burstStart = 50

func init() {
	burstFrames = frames.InitAbilSlice(63) // Q -> D/J
	burstFrames[action.ActionAttack] = 62  // Q -> N1
	burstFrames[action.ActionSkill] = 62   // Q -> E
	burstFrames[action.ActionSwap] = 62    // Q -> Swap
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Gyoei Narukami Kariyama Rite",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       0,
		FlatDmg:    c.MaxHP() * burst[c.TalentLvlBurst()],
	}
	snap := c.Snapshot(&ai)

	count := 7 // 低HP時は11回になる可能性あり
	if c.CurrentHPRatio() <= 0.5 {
		count = 12
	}
	interval := 2 * 60 / 7

	// 1凸: 裁雷除悪のAoE範囲が50%増加。
	r := 4.0
	if c.Base.Cons >= 1 {
		r = 6
	}

	// 元素爆発の中心がターゲットに十分近い前提
	for i := burstStart; i < count*interval+burstStart; i += interval {
		c.Core.QueueAttackWithSnap(
			ai,
			snap,
			combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, r),
			i,
		)
	}

	c.ConsumeEnergy(4)
	c.SetCD(action.ActionBurst, 900) // 15s * 60

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionAttack], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}
