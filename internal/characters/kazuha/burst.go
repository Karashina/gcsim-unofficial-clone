package kazuha

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var burstFrames []int

const (
	burstHitmark   = 82
	burstFirstTick = 140
)

const burstStatus = "kazuha-q"

func init() {
	burstFrames = frames.InitAbilSlice(93) // Q -> J
	burstFrames[action.ActionAttack] = 92  // Q -> N1
	burstFrames[action.ActionSkill] = 92   // Q -> E
	burstFrames[action.ActionDash] = 92    // Q -> D
	burstFrames[action.ActionSwap] = 90    // Q -> Swap
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	player := c.Core.Combat.Player()
	c.qAbsorb = attributes.NoElement
	c.qAbsorbCheckLocation = combat.NewCircleHitOnTarget(player, geometry.Point{Y: 1}, 8)

	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Kazuha Slash",
		AttackTag:          attacks.AttackTagElementalBurst,
		ICDTag:             attacks.ICDTagNone,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeSlash,
		Element:            attributes.Anemo,
		Durability:         50,
		Mult:               burstSlash[c.TalentLvlBurst()],
		HitlagHaltFrames:   0.05 * 60,
		HitlagFactor:       0.05,
		CanBeDefenseHalted: false,
	}
	ap := combat.NewCircleHitOnTarget(player, geometry.Point{Y: 1}, 9)

	c.Core.QueueAttack(ai, ap, burstHitmark, burstHitmark)

	// 元素吸収チェックとドットダメージを適用
	ai.Abil = "Kazuha Slash (Dot)"
	ai.StrikeType = attacks.StrikeTypeDefault
	ai.Mult = burstDot[c.TalentLvlBurst()]
	ai.Durability = 25
	// 初撃ヒットラグ以降はヒットラグなし
	ai.HitlagHaltFrames = 0

	aiAbsorb := ai
	aiAbsorb.Abil = "Kazuha Slash (Absorb Dot)"
	aiAbsorb.Mult = burstEleDot[c.TalentLvlBurst()]
	aiAbsorb.Element = attributes.NoElement

	c.Core.Tasks.Add(c.absorbCheckQ(c.Core.F, 0, int(310/18)), burstHitmark-1)

	// 2凸を処理
	// 最初のティックは初撃直前、0.5秒ごとに元素爆発中ティック
	c.QueueCharTask(func() {
		// スラッシュ直前にティックをスナップショット
		c.qTickSnap = c.Snapshot(&ai)
		c.qTickAbsorbSnap = c.Snapshot(&aiAbsorb)

		c.Core.Status.Add(burstStatus, (burstFirstTick-(burstHitmark-1))+117*4)
		if c.Base.Cons >= 2 {
			c.qFieldSrc = c.Core.F
			c.c2(c.Core.F)() // 即座にティックを開始
		}
	}, burstHitmark-1)

	// このタスクが実行されることを保証:
	// - 元素爆発のヒットラグ内
	// - 他のヒットラグの影響を受ける前
	c.QueueCharTask(func() {
		// ティックをキューに追加
		// kisaのカウントによると: ティックは147fから開始、約117f間隔で合計5ティック
		// koliのカウントに基づいて140に更新: https://docs.google.com/spreadsheets/d/1uEbP13O548-w_nGxFPGsf5jqj1qGD3pqFZ_AiV4w3ww/edit#gid=775340159
		for i := 0; i < 5; i++ {
			c.Core.Tasks.Add(func() {
				if c.qAbsorb != attributes.NoElement {
					aiAbsorb.Element = c.qAbsorb
					c.Core.QueueAttackWithSnap(aiAbsorb, c.qTickAbsorbSnap, ap, 0)
				}
				c.Core.QueueAttackWithSnap(ai, c.qTickSnap, ap, 0)
			}, (burstFirstTick-(burstHitmark+1))+117*i)
		}
		// 6凸:
		// TODO: 元素付与はいつ発動する？
		// -> 現状、初撃ヒットラグ終了時に開始すると仮定。
		if c.Base.Cons >= 6 {
			c.c6()
		}
	}, burstHitmark+1)

	// スキルクールダウンをリセット
	if c.Base.Cons >= 1 {
		c.ResetActionCooldown(action.ActionSkill)
	}

	c.SetCD(action.ActionBurst, 15*60)
	c.ConsumeEnergy(4)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}

func (c *char) absorbCheckQ(src, count, maxcount int) func() {
	return func() {
		if count == maxcount {
			return
		}
		c.qAbsorb = c.Core.Combat.AbsorbCheck(c.Index, c.qAbsorbCheckLocation, attributes.Pyro, attributes.Hydro, attributes.Electro, attributes.Cryo)

		if c.qAbsorb != attributes.NoElement {
			return
		}
		// それ以外はキューに追加
		c.Core.Tasks.Add(c.absorbCheckQ(src, count+1, maxcount), 18)
	}
}
