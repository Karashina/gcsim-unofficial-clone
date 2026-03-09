package klee

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var burstFrames []int
var waveHitmarks = []int{186, 294, 401, 503, 610, 718}

const burstStart = 146

func init() {
	burstFrames = frames.InitAbilSlice(139) // Q -> N1/CA/E
	burstFrames[action.ActionDash] = 103    // Q -> D
	burstFrames[action.ActionJump] = 104    // Q -> J
	burstFrames[action.ActionWalk] = 102    // Q -> Walk
	burstFrames[action.ActionSwap] = 101    // Q -> Swap
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Sparks'n'Splash",
		AttackTag:          attacks.AttackTagElementalBurst,
		ICDTag:             attacks.ICDTagElementalBurst,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeDefault,
		Element:            attributes.Pyro,
		Durability:         25,
		Mult:               burst[c.TalentLvlBurst()],
		NoImpulse:          true,
		CanBeDefenseHalted: true,
		IsDeployable:       true,
	}
	// 10秒間持続、2.2秒後に開始？
	c.Core.Status.Add("kleeq", 600+burstStart)

	// 1.8秒ごとに3～5発発射。キュー処理は無視。各ウェーブ間は0.2秒間隔

	// アニメーション終了時にスナップショット？
	var snap combat.Snapshot
	c.Core.Tasks.Add(func() {
		snap = c.Snapshot(&ai)
	}, 100)

	for _, start := range waveHitmarks {
		c.Core.Tasks.Add(func() {
			// 元素爆発が早期終了した場合は停止
			if c.Core.Status.Duration("kleeq") <= 0 {
				return
			}
			// ウェーブ1 = 1発
			c.Core.QueueAttackWithSnap(ai, snap, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 1.5), 0)
			// ウェーブ2 = 1発 + 30%の確率で追加1発
			c.Core.QueueAttackWithSnap(ai, snap, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 1.5), 12)
			if c.Core.Rand.Float64() < 0.3 {
				c.Core.QueueAttackWithSnap(ai, snap, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 1.5), 12)
			}
			// ウェーブ3 = 1発 + 50%の確率で追加1発
			c.Core.QueueAttackWithSnap(ai, snap, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 1.5), 24)
			if c.Core.Rand.Float64() < 0.5 {
				c.Core.QueueAttackWithSnap(ai, snap, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 1.5), 24)
			}
		}, start)
	}

	// 6凸の場合、3秒ごとにエネルギーを追加
	if c.Base.Cons >= 6 {
		//TODO: 最終的にはヒットラグ影響付きキューと持続時間を使用すべき
		// ただしクレーは被弾しないとヒットラグが発生しないため現時点では大きな問題ではない
		for i := burstStart + 180; i < burstStart+600; i += 180 {
			c.Core.Tasks.Add(func() {
				// 元素爆発が早期終了した場合は停止
				if c.Core.Status.Duration("kleeq") <= 0 {
					return
				}

				for i, x := range c.Core.Player.Chars() {
					if i == c.Index {
						continue
					}
					x.AddEnergy("klee-c6", 3)
				}
			}, i)
		}

		// 25秒間炎元素ダメージ+10%
		m := make([]float64, attributes.EndStatType)
		m[attributes.PyroP] = .1
		for _, x := range c.Core.Player.Chars() {
			x.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag("klee-c6", 1500),
				AffectedStat: attributes.PyroP,
				Amount: func() ([]float64, bool) {
					return m, true
				},
			})
		}
	}

	c.c1(waveHitmarks[0])

	c.SetCDWithDelay(action.ActionBurst, 15*60, 9)
	c.ConsumeEnergy(12)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // 最早キャンセルフレーム
		State:           action.BurstState,
	}, nil
}

// クレーがフィールドを離れた時に元素爆発を解除し、4凸を処理する
func (c *char) onExitField() {
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(_ ...interface{}) bool {
		// 元素爆発がアクティブかチェック
		if c.Core.Status.Duration("kleeq") <= 0 {
			return false
		}
		c.Core.Status.Delete("kleeq")

		if c.Base.Cons >= 4 {
			// 爆発
			ai := combat.AttackInfo{
				ActorIndex:         c.Index,
				Abil:               "Sparks'n'Splash C4",
				AttackTag:          attacks.AttackTagNone,
				ICDTag:             attacks.ICDTagNone,
				ICDGroup:           attacks.ICDGroupDefault,
				StrikeType:         attacks.StrikeTypeDefault,
				Element:            attributes.Pyro,
				Durability:         50,
				Mult:               5.55,
				CanBeDefenseHalted: true,
				IsDeployable:       true,
			}
			c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5), 0, 0)
		}

		return false
	}, "klee-exit")
}
