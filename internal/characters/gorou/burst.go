package gorou

import (
	"math"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/reactable"
)

var burstFrames []int

const burstHitmark = 31 // 初撃

func init() {
	burstFrames = frames.InitAbilSlice(56) // Q -> E
	burstFrames[action.ActionAttack] = 53  // Q -> N1
	burstFrames[action.ActionDash] = 42    // Q -> D
	burstFrames[action.ActionJump] = 43    // Q -> J
	burstFrames[action.ActionSwap] = 55    // Q -> Swap
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	// 初撃
	// A1/C6/Qの持続時間は全て初撃時に開始
	c.Core.Tasks.Add(func() {
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Juuga: Forward Unto Victory",
			AttackTag:  attacks.AttackTagElementalBurst,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeBlunt,
			PoiseDMG:   40,
			Element:    attributes.Geo,
			Durability: 25,
			Mult:       burst[c.TalentLvlBurst()],
			FlatDmg:    c.a4Burst(),
			UseDef:     true,
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5), 0, 0)

		// Q General's Glory（獣牙突撃陣形）:
		// 犬坂鐌繰の昭で作成される「大将旗」と同様に、スキルAoE内の
		// 岩元素キャラ数に応じたバフをアクティブキャラに付与。
		// アクティブキャラと共に移動する。
		c.eFieldSrc = c.Core.F
		c.Core.Tasks.Add(c.gorouSkillBuffField(c.Core.F), 17) // 17にして最後のTickを取得

		// このアビリティ使用時にゴローが作成した「大将旗」がフィールド上に存在する場合、
		// それは破壊される。さらに、General's Gloryの持続中、ゴローの
		// 元素スキル「犬坂鐌繰の昭」は「大将旗」を作成しない。
		c.Core.Status.Delete(generalWarBannerKey)
		c.Core.Status.Add(generalGloryKey, generalGloryDuration) // フィールドはヒットマーク初撃から開始

		// 1.5秒ごとに1つのCrystal Collapseを生成し、スキルAoE内の敵1体にAoE岩元素ダメージ。
		// 1.5秒ごとにスキルAoE内の元素結晶片を1つアクティブキャラの位置に引き寄せる
		// （元素結晶片は結晶反応で生成される）。
		c.qFieldSrc = c.Core.F
		c.Core.Tasks.Add(c.gorouCrystalCollapse(c.Core.F), 90) // 最初のCrystal Collapseは初撃ヒットマーク後1.5秒

		c.a1()

		// 4凸
		if c.Base.Cons >= 4 && c.geoCharCount > 1 {
			// TODO: 実際にステータスをスナップショットするか不明...
			// ai := combat.AttackInfo{
			// 	Abil:      "Inuzaka All-Round Defense C4",
			// 	AttackTag: attacks.AttackTagNone,
			// }
			c.healFieldStats, _ = c.Stats()
			c.Core.Tasks.Add(c.gorouBurstHealField(c.Core.F), 90)
		}

		// 6凸
		if c.Base.Cons >= 6 {
			c.c6()
		}
	}, burstHitmark)

	//TODO: ゴローが倒れると、General's Gloryの効果は解除される。

	c.c2Extension = 0

	c.SetCD(action.ActionBurst, 20*60)
	c.ConsumeEnergy(7)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}

// ダメージを与える再帰関数
func (c *char) gorouCrystalCollapse(src int) func() {
	return func() {
		// 上書きされている場合は何もしない
		if c.qFieldSrc != src {
			return
		}
		// フィールドが期限切れなら何もしない
		if c.Core.Status.Duration(generalGloryKey) == 0 {
			return
		}
		// ダメージを発動
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Crystal Collapse",
			AttackTag:  attacks.AttackTagElementalBurst,
			ICDTag:     attacks.ICDTagElementalBurst,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Geo,
			Durability: 25,
			Mult:       burstTick[c.TalentLvlBurst()],
			FlatDmg:    c.a4Burst(),
			UseDef:     true,
		}
		collapseArea := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 8)
		enemy := c.Core.Combat.ClosestEnemyWithinArea(collapseArea, nil)
		if enemy != nil {
			//TODO: スキルダメージのフレーム
			c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(enemy, nil, 3.5), 0, 1)
		}

		// 元素結晶片を1つ吸引
		for _, g := range c.Core.Combat.Gadgets() {
			cs, ok := g.(*reactable.CrystallizeShard)
			// 結晶がなければスキップ
			if !ok {
				continue
			}
			// 結晶片が範囲外ならスキップ
			if !cs.IsWithinArea(collapseArea) {
				continue
			}
			// 吸引を0.4m/フレームとして近似（約8mの距離がゴローまで20fで到着）
			distance := cs.Pos().Distance(collapseArea.Shape.Pos())
			travel := int(math.Ceil(distance / 0.4))
			// 結晶が生成直後で拾える前に到着するエッジケースのための特別チェック
			if c.Core.F+travel < cs.EarliestPickup {
				continue
			}
			c.Core.Tasks.Add(func() {
				cs.AddShieldKillShard()
			}, travel)
			break
		}

		// 1.5秒ごとにTick
		c.Core.Tasks.Add(c.gorouCrystalCollapse(src), 90)
	}
}

func (c *char) gorouBurstHealField(src int) func() {
	return func() {
		// 上書きされている場合は何もしない
		if c.qFieldSrc != src {
			return
		}
		// フィールドが期限切れなら何もしない
		if c.Core.Status.Duration(generalGloryKey) == 0 {
			return
		}
		// 「集岩」または「碎岩」状態のGeneral's Gloryは、AoE内のアクティブキャラを
		// 1.5秒ごとにゴロー自身の防御力の50%分回復する。
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index,
			Target:  c.Core.Player.Active(),
			Message: "Lapping Hound: Warm as Water",
			Src:     c.healFieldStats.TotalDEF() * 0.5,
			Bonus:   c.Stat(attributes.Heal),
		})

		// 1.5秒ごとにTick
		c.Core.Tasks.Add(c.gorouBurstHealField(src), 90)
	}
}
