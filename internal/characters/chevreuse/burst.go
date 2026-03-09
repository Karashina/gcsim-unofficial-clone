package chevreuse

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var (
	burstFrames []int
)

func init() {
	burstFrames = frames.InitAbilSlice(61) // Q -> Walk
	burstFrames[action.ActionAttack] = 57
	burstFrames[action.ActionSkill] = 59
	burstFrames[action.ActionDash] = 57
	burstFrames[action.ActionJump] = 57
	burstFrames[action.ActionSwap] = 56
}

const (
	burstHitmark  = 59
	snapshotDelay = 43
)

func (c *char) Burst(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Explosive Grenade",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		Element:    attributes.Pyro,
		PoiseDMG:   100,
		Durability: 50,
		Mult:       burst[c.TalentLvlBurst()],
	}

	mineAi := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Secondary Explosive Shell",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupChevreuseBurstMines,
		StrikeType: attacks.StrikeTypeBlunt,
		PoiseDMG:   25,
		Element:    attributes.Pyro,
		Durability: 25,
		Mult:       burstSecondary[c.TalentLvlBurst()],
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHit(c.Core.Combat.Player(), c.Core.Combat.PrimaryTarget(), nil, 6),
		snapshotDelay,
		burstHitmark,
	)

	burstInitialDirection := c.Core.Combat.Player().Direction()
	burstInitialPos := c.Core.Combat.PrimaryTarget().Pos()
	// 地雷は合計8個、グループごとに爆発
	// 5グループの地雷
	// 基本的に:
	// - 円を8つのスライスに分割
	// - 上部の地雷から爆発開始
	// - 各半分と1つずつ2つの地雷を爆発させ、下部の地雷に到達するまで続ける
	// - 下部の地雷が最後に爆発（プレイヤーに最も近い）
	mineGroups := 5
	mineCounts := []int{1, 2, 2, 2, 1}
	mineSteps := [][]float64{{0}, {45, 315}, {90, 270}, {135, 225}, {180}}
	mineDelays := []int{24, 33, 42, 51, 60}
	for i := 0; i < mineGroups; i++ {
		for j := 0; j < mineCounts[i]; j++ {
			// 各爆弾はそれぞれ独自の方向を持つ
			direction := geometry.DegreesToDirection(mineSteps[i][j]).Rotate(burstInitialDirection)

			// 方向を簡単に指定できないため combat の攻撃パターン関数を使用できない
			mineAp := combat.AttackPattern{
				Shape: geometry.NewCircle(burstInitialPos, 6, direction, 60),
			}
			mineAp.SkipTargets[targets.TargettablePlayer] = true
			c.Core.QueueAttack(mineAi, mineAp, snapshotDelay, burstHitmark+mineDelays[i])
		}
	}

	c.c4()
	c.ConsumeEnergy(4)
	c.SetCD(action.ActionBurst, 15*60)
	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}
