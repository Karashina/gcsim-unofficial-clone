package navia

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var (
	burstFrames []int
)

const (
	burstHitmark  = 104
	burstKey      = "navia-artillery"
	burstDuration = 720
	burstDelay    = 154
	burstICDKey   = "navia-q-shrapnel-icd"
)

func init() {
	burstFrames = frames.InitAbilSlice(127)
	burstFrames[action.ActionAttack] = 102
	burstFrames[action.ActionSkill] = 102
	burstFrames[action.ActionDash] = 103
	burstFrames[action.ActionJump] = 103
	burstFrames[action.ActionSwap] = 93
}

// 薔薇の会会長の命令により、壮大なRosula Dorata Saluteを発動。
// 前方の敵に大規模な砲撃を行い、岩元素範囲ダメージを与え、
// 一定時間砲撃支援を提供し、定期的に付近の敵に岩元素ダメージを与える。
func (c *char) Burst(_ map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "As the Sunlit Sky's Singing Salute",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		PoiseDMG:   100,
		Element:    attributes.Geo,
		Durability: 50,
		Mult:       burst[0][c.TalentLvlBurst()],
	}

	c.Core.QueueAttack(
		ai,
		combat.NewBoxHitOnTarget(c.Core.Combat.Player(), nil, 5, 12),
		burstHitmark,
		burstHitmark,
		c.burstCB(),
		c.c4(),
	)

	c.QueueCharTask(func() {
		c.AddStatus(burstKey, burstDuration, false)

		ai.Abil = "Cannon Fire Support"
		ai.ICDTag = attacks.ICDTagElementalBurst
		ai.ICDGroup = attacks.ICDGroupNaviaBurst
		ai.PoiseDMG = 50
		ai.Durability = 25
		ai.Mult = burst[1][c.TalentLvlBurst()]

		tick := 0
		var nextTick int
		for i := 0; i <= burstDuration; i += nextTick {
			tick++
			c.Core.Tasks.Add(func() {
				// 攻撃をキューに追加
				c.Core.QueueAttack(
					ai,
					combat.NewCircleHitOnTarget(c.calcCannonPos(), nil, 3),
					0,
					9,
					c.burstCB(),
					c.c4(),
				)
			}, i)
			// tick 2, 5, 8, 11, 14 がキューされた場合、次のティックは42fではなく48f後
			if tick%3 == 2 {
				nextTick = 48
			} else {
				nextTick = 42
			}
		}
	}, burstDelay)

	c.ConsumeEnergy(12)
	c.SetCD(action.ActionBurst, 15*60)
	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}

// 砲撃が敵に命中すると、ナヴィアはCrystal Shrapnelを1スタック獲得。
// この効果は2.4秒ごとに1回まで発動可能。
func (c *char) burstCB() combat.AttackCBFunc {
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if c.StatusIsActive(burstICDKey) {
			return
		}
		c.AddStatus(burstICDKey, 2.4*60, true)
		if c.shrapnel < 6 {
			c.shrapnel++
			c.Core.Log.NewEvent("Crystal Shrapnel gained from Burst", glog.LogCharacterEvent, c.Index).Write("shrapnel", c.shrapnel)
		}
	}
}

// 敵がいればランダムな敵を、いなければランダムな地点をターゲットする
func (c *char) calcCannonPos() geometry.Point {
	player := c.Core.Combat.Player() // ガジェットはプレイヤーに付属

	// プレイヤー位置から10m半径内のランダムな敵を検索
	enemy := c.Core.Combat.RandomEnemyWithinArea(
		combat.NewCircleHitOnTarget(c.Core.Combat.Player().Pos(), nil, 10),
		nil,
	)

	// 敵が見つかった: 敵の位置から0～1.2mのランダムな点を選択
	if enemy != nil {
		return geometry.CalcRandomPointFromCenter(enemy.Pos(), 0, 1.2, c.Core.Rand)
	}

	// 敵なし: プレイヤー位置 + Y:4 から1m～6mのランダムな地点をターゲット
	return geometry.CalcRandomPointFromCenter(
		geometry.CalcOffsetPoint(
			player.Pos(),
			geometry.Point{Y: 4},
			player.Direction(),
		),
		1,
		6,
		c.Core.Rand,
	)
}
