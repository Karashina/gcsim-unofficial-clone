package lisa

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var burstFrames []int

const burstHitmark = 56

func init() {
	burstFrames = frames.InitAbilSlice(88)
	burstFrames[action.ActionAttack] = 86
	burstFrames[action.ActionCharge] = 86
	burstFrames[action.ActionSkill] = 87
	burstFrames[action.ActionJump] = 57
	burstFrames[action.ActionSwap] = 56
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	// 最初の電撃はICDなしで全員に命中
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Lightning Rose (Initial)",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 0,
		Mult:       0.1,
	}
	// nosiとの議論により、これは防御低下を適用しないことが判明
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 7), burstHitmark, burstHitmark)

	// 持続時間15秒、0.5秒ごとにTick
	// 30フレームごとに30回電撃、119フレーム目から開始
	ai = combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Lightning Rose (Tick)",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       burst[c.TalentLvlBurst()],
	}

	var snap combat.Snapshot
	c.Core.Tasks.Add(func() {
		snap = c.Snapshot(&ai)
	}, burstHitmark-1)

	firstTick := 119 // 最初のティックは119フレーム目
	burstArea := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 7)
	for i := 0; i < 15*60; i += 30 {
		progress := i + firstTick
		c.Core.Tasks.Add(func() {
			// 4凸未満のロジックは単純: 範囲内のランダムな敵に1回放電
			if c.Base.Cons < 4 {
				enemy := c.Core.Combat.RandomEnemyWithinArea(burstArea, nil)
				if enemy == nil {
					return
				}
				c.Core.QueueAttackWithSnap(
					ai,
					snap,
					combat.NewCircleHitOnTarget(enemy, nil, 1),
					0,
					c.makeA4CB(),
				)
				return
			}

			// 4凸以上:
			// - https://library.keqingmains.com/evidence/characters/electro/lisa#c4-plasma-eruption
			// - 敵+ガジェットの数に基づき最大3回攻撃を生成
			// - 優先度: 敵 > ガジェット
			discharge := func(pos geometry.Point) {
				c.Core.QueueAttackWithSnap(
					ai,
					snap,
					combat.NewCircleHitOnTarget(pos, nil, 1),
					0,
					c.makeA4CB(),
				)
			}
			dischargeCount := 0
			dischargeLimit := 3

			enemies := c.Core.Combat.RandomEnemiesWithinArea(burstArea, nil, dischargeLimit)
			enemyCount := len(enemies)

			var gadgets []combat.Gadget
			if enemyCount < dischargeLimit {
				gadgets = c.Core.Combat.RandomGadgetsWithinArea(burstArea, nil, dischargeLimit-enemyCount)
			}
			gadgetCount := len(gadgets)

			totalEntities := enemyCount + gadgetCount
			switch totalEntities {
			case 0:
			case 1:
				dischargeCount = 1
			case 2:
				threshold := 0.15
				if progress == firstTick {
					// 最初の放電: 60%の確率で1回、40%の確率で2回
					threshold = 0.6
				}
				// 残り: 15%の確率で1回、85%の確率で2回
				if c.Core.Rand.Float64() < threshold {
					dischargeCount = 1
				} else {
					dischargeCount = 2
				}
			default: // 3 or more entities
				if progress == firstTick {
					// 初回: 55%で1回、45%で2回
					if c.Core.Rand.Float64() < 0.55 {
						dischargeCount = 1
					} else {
						dischargeCount = 2
					}
					c.previousDischargeCount = dischargeCount
					if dischargeCount == 0 {
						return
					}
					return
				}
				rand := c.Core.Rand.Float64()
				switch c.previousDischargeCount {
				case 1:
					// 1回の後: 20%で1回、50%で2回、30%で3回
					switch {
					case rand < 0.2:
						dischargeCount = 1
					case rand < 0.7:
						dischargeCount = 2
					default:
						dischargeCount = 3
					}
				case 2:
					// 2回の後: 25%で1回、50%で2回、25%で3回
					switch {
					case rand < 0.25:
						dischargeCount = 1
					case rand < 0.75:
						dischargeCount = 2
					default:
						dischargeCount = 3
					}
				case 3:
					// 3回の後: 50%で1回、50%で2回、0%で3回
					if rand < 0.5 {
						dischargeCount = 1
					} else {
						dischargeCount = 2
					}
				}
			}
			c.previousDischargeCount = dischargeCount
			if dischargeCount == 0 {
				return
			}

			// 最大3体の敵をターゲット
			for i := 0; i < dischargeCount; i++ {
				if i < enemyCount {
					discharge(enemies[i].Pos())
				}
			}
			dischargeCount -= enemyCount

			// 3体未満の敵がターゲットされた場合、ガジェットをチェック
			for i := 0; i < dischargeCount; i++ {
				if i < gadgetCount {
					discharge(gadgets[i].Pos())
				}
			}
		}, progress)
	}

	// 念のためステータスを追加
	c.Core.Tasks.Add(func() {
		c.Core.Status.Add("lisaburst", 119+900)
	}, burstHitmark)

	// 元素爆発のCDは実行後53フレームで開始
	// エネルギーは通常63フレーム後に消費
	c.ConsumeEnergy(63)
	// c.CD[def.BurstCD] = c.Core.F + 1200
	c.SetCDWithDelay(action.ActionBurst, 1200, 53)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}
