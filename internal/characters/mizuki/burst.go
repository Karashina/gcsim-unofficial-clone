package mizuki

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
	burstActivationDmgName     = "Anraku Secret Spring Therapy"
	burstKey                   = "mizuki-burst"
	burstHitmark               = 93
	burstDurability            = 50
	burstPoise                 = 100
	burstDuration              = 12 * 60
	burstCdDelay               = 1
	burstEnergyDrainDelay      = 4
	burstCd                    = 15 * 60
	burstRadius                = 8
	snackInterval              = 1.5 * 60
	snackSpawnOnEnemyRadius    = 6
	snackSpawnLocationVariance = 1.0
)

func init() {
	burstFrames = frames.InitAbilSlice(94) // Q -> Swap
	burstFrames[action.ActionAttack] = 93
	burstFrames[action.ActionCharge] = 92
	burstFrames[action.ActionSkill] = 93
	burstFrames[action.ActionDash] = 91
	burstFrames[action.ActionJump] = 93
	burstFrames[action.ActionWalk] = 92
}

// 数え切れないほどの美しい夢と悪夢を呼び出し、周囲のオブジェクトと敵を引き寄せ、
// 範囲風元素ダメージを与えてミニバクを召喚する。
func (c *char) Burst(p map[string]int) (action.Info, error) {
	// 発動ダメージ
	ai := combat.AttackInfo{
		ActorIndex:   c.Index,
		Abil:         burstActivationDmgName,
		AttackTag:    attacks.AttackTagElementalBurst,
		ICDTag:       attacks.ICDTagNone,
		ICDGroup:     attacks.ICDGroupDefault,
		StrikeType:   attacks.StrikeTypeDefault,
		Element:      attributes.Anemo,
		Durability:   burstDurability,
		PoiseDMG:     burstPoise,
		Mult:         burstDMG[c.TalentLvlBurst()],
		HitlagFactor: 0.05,
	}

	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, burstRadius), burstHitmark, burstHitmark)

	// スクリプトでの確認に役立つ可能性あり
	c.AddStatus(burstKey, burstDuration, false)

	if c.Base.Cons >= 4 {
		c.c4EnergyGenerationsRemaining = c4EnergyGenerations
	}

	c.queueSnacks()

	c.ConsumeEnergy(burstEnergyDrainDelay)
	c.SetCDWithDelay(action.ActionBurst, burstCd, burstCdDelay)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash],
		State:           action.BurstState,
	}, nil
}

func (c *char) queueSnacks() {
	randomSign := func() float64 {
		rnd := c.Core.Rand.Float64()
		if 0.5 < rnd {
			return -1
		}
		return 1
	}
	snackFunc := func() {
		pos := c.calculateSnackSpawnLocation()

		// スポーン位置を少しランダム化
		pos.Y += c.Core.Rand.Float64() * snackSpawnLocationVariance * randomSign()
		pos.X += c.Core.Rand.Float64() * snackSpawnLocationVariance * randomSign()

		newSnack(c, pos)
	}

	// スポーンタイマーは元素爆発のヒットマークから開始
	spawnTime := burstHitmark
	for i := int(snackInterval); i <= burstDuration; i += snackInterval {
		c.Core.Tasks.Add(snackFunc, spawnTime+i)
	}
}

func (c *char) calculateSnackSpawnLocation() geometry.Point {
	// テストによると、おやつはターゲット/プレイヤーの前方の狭い範囲（1m）に出現する。
	// ただし敵の方向はデフォルトでプレイヤーに向いていないため、
	// プレイヤー/敵からの相対位置を計算する
	playerPos := c.Core.Combat.Player().Pos()
	finalPosition := playerPos

	// 最も近い敵を探す
	target := c.Core.Combat.ClosestEnemyWithinArea(
		combat.NewCircleHitOnTarget(playerPos, nil, snackSpawnOnEnemyRadius),
		nil,
	)

	// 敵が見つかればそれを使用、見つからなければプレイヤー位置を使用
	if target != nil {
		targetShape := target.Shape()
		finalPosition = targetShape.Pos()
		direction := geometry.Point{
			X: playerPos.X - finalPosition.X,
			Y: playerPos.Y - finalPosition.Y,
		}
		if v, ok := targetShape.(*geometry.Circle); ok {
			if finalPosition != playerPos {
				direction = direction.Normalize()
				finalPosition.X += v.Radius() * direction.X
				finalPosition.Y += v.Radius() * direction.Y
			}
		} else if _, ok := targetShape.(*geometry.Rectangle); ok {
			// 現在、矩形上でおやつをスポーンさせるエッジを確実に取得できない。
			// 中央付近に配置する
			finalPosition.X += direction.X / 2
			finalPosition.Y += direction.Y / 2
		}
	}
	return finalPosition
}
