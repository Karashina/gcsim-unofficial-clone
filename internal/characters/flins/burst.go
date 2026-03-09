package flins

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var (
	burstFrames      []int
	tsFrames         []int
	NascentHitmark   = []int{111, 125, 185}           // 中間フェーズ x2 + 最終フェーズ x1
	AscendantHitmark = []int{111, 125, 129, 138, 164} // 中間フェーズ x4 + 最終フェーズ x1
)

const (
	initialHitmark = 96
	tsHitmark      = 45
	tsAddHitmark   = 66
)

func init() {
	burstFrames = frames.InitAbilSlice(113)
	tsFrames = frames.InitAbilSlice(55)
}

// Q
// Ancient Ritual: Cometh the Night（元素爆発）
// Flinsが単体のAoE雷元素ダメージを与え、短い遅延の後、2回の中間フェーズと1回の最終フェーズAoE雷元素ダメージを与える。これらはすべてルナチャージダメージとみなされる。
// ムーンサインが「ムーンサイン: 昇詼の輝き」の場合、このアビリティは強化される: 付近に雷雲がある場合、Flinsは中間フェーズのルナチャージAoE雷元素ダメージを追加2回与える。

func (c *char) Burst(p map[string]int) (action.Info, error) {
	// 特殊元素スキル「北地の槍嵐」使用後、Flinsの元素爆発「Ancient Ritual: Cometh the Night」は次の6秒間特殊元素爆発「雷鳴の交響曲」に置き換わる。
	// 雷鳴の交響曲
	// より少ない元素エネルギーを消費して特殊元素爆発を発動。Flinsが単体のAoE雷元素ダメージを与える（ルナチャージダメージとみなされる）。
	// ムーンサインが「ムーンサイン: 昇詼の輝き」の場合、Flinsの元素スキルは強化される: 付近に雷雲がある場合、FlinsはルナチャージAoE雷元素ダメージを追加1回与える。
	if c.StatusIsActive(northlandKey) {
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Flins TS Dummy",
			FlatDmg:    0,
		}
		aiadd := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Flins TSADD Dummy",
			FlatDmg:    0,
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 99), tsHitmark, tsHitmark)
		if c.MoonsignAscendant && c.lcCloudCheck() {
			c.Core.QueueAttack(aiadd, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 99), tsAddHitmark, tsAddHitmark)
		}

		c.DeleteStatus(northlandKey)
		c.ConsumeEnergyPartial(7, 30)

		return action.Info{
			Frames:          frames.NewAbilFunc(tsFrames),
			AnimationLength: tsFrames[action.InvalidAction],
			CanQueueAfter:   tsFrames[action.ActionSwap], // 最速キャンセル
			State:           action.BurstState,
		}, nil
	} else {
		// 初撃
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Initial Skill DMG (Q)",
			AttackTag:  attacks.AttackTagElementalBurst,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Electro,
			Durability: 25,
			Mult:       burst[c.TalentLvlBurst()],
		}
		c.QueueCharTask(func() {
			c.Core.QueueAttack(
				ai,
				combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: -1.5}, 7),
				0, 0,
			)
		}, initialHitmark)

		aimid := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Flins QMid Dummy",
			FlatDmg:    0,
		}
		aifin := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Flins QFin Dummy",
			FlatDmg:    0,
		}
		if c.MoonsignAscendant && c.lcCloudCheck() {
			// 中間フェーズ x4 + 最終フェーズ x1
			for i, hitmark := range AscendantHitmark {
				if i < 4 {
					// 中間フェーズ
					c.Core.QueueAttack(aimid, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 99), hitmark, hitmark)
				} else {
					// 最終フェーズ
					c.Core.QueueAttack(aifin, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 99), hitmark, hitmark)
				}
			}
		} else {
			// 中間フェーズ x2 + 最終フェーズ x1
			for i, hitmark := range NascentHitmark {
				if i < 2 {
					// 中間フェーズ
					c.Core.QueueAttack(aimid, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 99), hitmark, hitmark)
				} else {
					// 最終フェーズ
					c.Core.QueueAttack(aifin, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 99), hitmark, hitmark)
				}
			}
		}
		c.SetCD(action.ActionBurst, 20*60)
		c.ConsumeEnergy(7)

		return action.Info{
			Frames:          frames.NewAbilFunc(burstFrames),
			AnimationLength: burstFrames[action.InvalidAction],
			CanQueueAfter:   burstFrames[action.ActionSwap], // 最速キャンセル
			State:           action.BurstState,
		}, nil
	}
}

func (c *char) lcCloudCheck() bool {
	for _, target := range c.Core.Combat.Enemies() {
		if c.HasLCCloudOn(target) {
			return true
		}
	}
	return false
}
