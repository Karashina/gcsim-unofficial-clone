package rosaria

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var burstFrames []int

func init() {
	burstFrames = frames.InitAbilSlice(70)
	burstFrames[action.ActionDash] = 57
	burstFrames[action.ActionJump] = 59
	burstFrames[action.ActionSwap] = 69
}

// 元素爆発のダメージキュー生成
// ロサリアは武器を振るい周囲の敵を斜り、冷たいIce Lanceを召喚して地面に突き刺す。両方とも氷元素ダメージを与える。
// 活性化中、Ice Lanceは定期的に冷気を放出し、周囲の敵に氷元素ダメージを与える。
// 固有天賦4、6凸の効果も含む
func (c *char) Burst(p map[string]int) (action.Info, error) {
	// 注意: より高度なターゲッティングシステムが将来追加された場合、
	// 1撃目は技術的には周囲の敵のみ、第2撃とDoTは槍に当たる
	// 現時点では全てが全ターゲットに当たると仮定
	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Rites of Termination (Hit 1)",
		AttackTag:          attacks.AttackTagElementalBurst,
		ICDTag:             attacks.ICDTagNone,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeSlash,
		Element:            attributes.Cryo,
		Durability:         25,
		Mult:               burst[0][c.TalentLvlBurst()],
		HitlagHaltFrames:   0.06 * 60,
		HitlagFactor:       0.01,
		CanBeDefenseHalted: false,
	}

	c1CB := c.makeC1CB()
	c6CB := c.makeC6CB()

	// 1撃目は15フレーム目に発生
	// 2撃目は槍の落下アニメーション後に発生
	// プレイヤー中心
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0.5}, 3.5),
		15,
		15,
		c1CB,
		c6CB,
	)

	ai.Abil = "Rites of Termination (Hit 2)"
	ai.StrikeType = attacks.StrikeTypeDefault
	ai.Mult = burst[1][c.TalentLvlBurst()]
	// 最初のヒット以降ヒットラグなし
	ai.HitlagHaltFrames = 0

	// 持続時間は8秒（c2で4秒延長）+ 0.5
	// 設置物であるべき
	dur := 510
	if c.Base.Cons >= 2 {
		dur += 240
	}

	playerPos := c.Core.Combat.Player()
	gadgetOffset := geometry.Point{Y: 3}
	apHit2 := combat.NewCircleHitOnTarget(playerPos, gadgetOffset, 6)
	apTick := combat.NewCircleHitOnTarget(playerPos, gadgetOffset, 6.5)
	// 2撃目とDoTの処理
	// 槍は56フレーム目に落下（ヒットラグ除外時。60フレームはヒットラグ含む）
	c.QueueCharTask(func() {
		// 2撃目
		c.Core.QueueAttack(ai, apHit2, 0, 0, c1CB, c6CB)

		// 元素爆発ステータス
		c.Core.Status.Add("rosariaburst", dur)

		// 元素爆発は槍が落下した時点（第2ダメージヒット時）にスナップショット
		ai = combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Rites of Termination (DoT)",
			AttackTag:  attacks.AttackTagElementalBurst,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			Element:    attributes.Cryo,
			Durability: 25,
			Mult:       burstDot[c.TalentLvlBurst()],
		}

		// 槍落下後、2秒ごとにDoT
		for i := 120; i < dur; i += 120 {
			c.Core.QueueAttack(ai, apTick, 0, i, c1CB, c6CB)
		}
	}, 56)

	c.a4()

	c.SetCD(action.ActionBurst, 15*60)
	c.ConsumeEnergy(6)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}
