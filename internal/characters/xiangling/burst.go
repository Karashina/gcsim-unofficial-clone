package xiangling

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
)

var (
	burstFrames   []int
	burstHitmarks = []int{18, 33, 57} // initial 3 hits
	burstRadius   = []float64{2.5, 2.5, 3}
)

func init() {
	burstFrames = frames.InitAbilSlice(80)
	burstFrames[action.ActionSwap] = 79
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	for i := range pyronadoInitial {
		initialHit := combat.AttackInfo{
			Abil:               fmt.Sprintf("Pyronado Hit %v", i+1),
			ActorIndex:         c.Index,
			AttackTag:          attacks.AttackTagElementalBurst,
			ICDTag:             attacks.ICDTagElementalBurst,
			ICDGroup:           attacks.ICDGroupDefault,
			StrikeType:         attacks.StrikeTypeDefault,
			Element:            attributes.Pyro,
			Durability:         25,
			HitlagHaltFrames:   0.03 * 60,
			HitlagFactor:       0.01,
			CanBeDefenseHalted: true,
			Mult:               pyronadoInitial[i][c.TalentLvlBurst()],
		}
		radius := burstRadius[i]
		c.QueueCharTask(func() {
			c.Core.QueueAttack(initialHit, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, radius), 0, 0)
		}, burstHitmarks[i])
	}

	// 約73フレームごとのサイクル
	// 最大10秒または14秒 + アニメーション
	// TODO: アニメーション長が正確かどうか不明
	a := 56

	burstHit := combat.AttackInfo{
		Abil:       "Pyronado",
		ActorIndex: c.Index,
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Pyro,
		Durability: 25,
		Mult:       pyronadoSpin[c.TalentLvlBurst()],
	}

	// 回転攻撃をaフレーム遅延させる。ヒットラグの影響を受けるべき
	c.QueueCharTask(func() {
		maxDuration := 10 * 60
		if c.Base.Cons >= 4 {
			maxDuration = 14 * 60
		}
		c.Core.Status.Add("xianglingburst", maxDuration)
		snap := c.Snapshot(&burstHit)
		for delay := 0; delay <= maxDuration; delay += 73 { // 最初のヒットは3回目の初撃の1f前
			// TODO: 適切なヒットボックス
			c.Core.Tasks.Add(func() {
				c.Core.QueueAttackWithSnap(
					burstHit,
					snap,
					combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 2.5),
					0,
				)
			}, delay)
		}
		// 6凸の場合、56フレーム目から持続時間終了まで炎元素ダメージ+15%の効果を追加
		if c.Base.Cons >= 6 {
			c.c6(maxDuration)
		}
	}, a)

	// シミュレーションにクールダウンを追加
	c.SetCDWithDelay(action.ActionBurst, 20*60, 18)
	// エネルギーを消費
	c.ConsumeEnergy(24)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}
