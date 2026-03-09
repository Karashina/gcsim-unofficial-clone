package albedo

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var burstFrames []int

const burstHitmark = 75                        // 初撃
const fatalBlossomHitmark = 145 - burstHitmark // Fatal Blossom、タスクキューイングを考慮

func init() {
	burstFrames = frames.InitAbilSlice(96) // Q -> N1/E
	burstFrames[action.ActionDash] = 95    // Q -> D
	burstFrames[action.ActionJump] = 94    // Q -> J
	burstFrames[action.ActionSwap] = 93    // Q -> Swap
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	// 6凸Hexerei追加効果: Silver Isotomaを破壊してFatal Blossomを強化
	if c.Base.Cons >= 6 && c.isHexerei {
		c.c6BlossomBuffOnBurst()
	}

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Rite of Progeniture: Tectonic Tide",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		PoiseDMG:   100,
		Element:    attributes.Geo,
		Durability: 25,
		Mult:       burst[c.TalentLvlBurst()],
	}

	c2Count := 0
	hasC2 := c.Base.Cons >= 2 && c.StatusIsActive(c2key)
	// 2凸の初撃ダメージは元素爆発開始時に計算される
	if hasC2 {
		c2Count = c.c2stacks
		c.c2stacks = 0
		ai.FlatDmg = c.TotalDef(false) * float64(c2Count) * 0.3
	}

	// 初撃ダメージ
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTargetFanAngle(c.Core.Combat.Player(), nil, 8, 120),
		burstHitmark,
		burstHitmark,
	)

	// 固有天賦4とFatal Blossom
	// Fatal Blossomの発動をburstHitmarkまで遅延させる。そのタイミングで:
	// - スキルがまだアクティブかチェックする
	// - 2凸ダメージを再計算する
	c.Core.Tasks.Add(func() {
		c.a4()

		// Fatal Blossom
		if !c.skillActive {
			return
		}
		ai.Abil = "Rite of Progeniture: Tectonic Tide (Blossom)"
		ai.PoiseDMG = 30
		ai.Mult = burstPerBloom[c.TalentLvlBurst()]

		// 2凸ダメージはburstHitmark時に1度だけ再計算される
		if hasC2 {
			ai.FlatDmg = c.TotalDef(false) * float64(c2Count) * 0.3
		}

		// 6凸Hexerei追加効果: Fatal Blossomのダメージに防御力の250%を加算
		if c.Base.Cons >= 6 && c.isHexerei && c.StatusIsActive(c6BlossomBuffKey) {
			ai.FlatDmg += c.TotalDef(false) * 2.5 // 防御力の250%
		}

		// 7個のBlossomを生成
		maxBlossoms := 7
		enemies := c.Core.Combat.RandomEnemiesWithinArea(c.skillArea, nil, maxBlossoms)
		tracking := len(enemies)
		var center geometry.Point
		for i := 0; i < maxBlossoms; i++ {
			if i < tracking {
				// 可能な限り各Blossomは別々の敵をターゲットする
				center = enemies[i].Pos()
			} else {
				// Blossomに敵がいない場合はスキル範囲内にランダムに出現
				center = geometry.CalcRandomPointFromCenter(c.skillArea.Shape.Pos(), 0.5, 9.5, c.Core.Rand)
			}
			// Blossomは初撃からわずかに遅れて生成される
			// TODO: Blossom間の正確なフレームデータなし
			c.Core.QueueAttackWithSnap(
				ai,
				c.bloomSnapshot,
				combat.NewCircleHitOnTarget(center, nil, 3),
				fatalBlossomHitmark+i*5,
			)
		}
	}, burstHitmark)

	c.SetCDWithDelay(action.ActionBurst, 720, 74)
	c.ConsumeEnergy(77)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap],
		State:           action.BurstState,
	}, nil
}
