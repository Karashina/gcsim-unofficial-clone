package kaeya

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
)

var burstFrames []int

const (
	burstCDStart  = 48
	burstHitmark  = 53
	burstDuration = 480
	burstKey      = "kaeya-q"
)

func init() {
	burstFrames = frames.InitAbilSlice(78) // Q -> E
	burstFrames[action.ActionAttack] = 77  // Q -> N1
	burstFrames[action.ActionDash] = 62    // Q -> D
	burstFrames[action.ActionJump] = 61    // Q -> J
	burstFrames[action.ActionSwap] = 77    // Q -> Swap
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Glacial Waltz",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Cryo,
		Durability: 25,
		Mult:       burst[c.TalentLvlBurst()],
	}
	snap := c.Snapshot(&ai)

	// 氷柱Tick用の元素爆発ステータスを追加
	// +1により12ではなく13Tick取得（ゲーム内では単体の敵に最大14/15回ヒット可能）
	c.Core.Status.Add(burstKey, burstDuration+burstHitmark+1)

	// 各氷柱は120フレームで一周し、0.5秒の内部クールダウンを持つ
	count := 3
	// 6凸:
	// 冰雪の輪は氷柱を1つ追加生成し、発動時に元素エネルギーを15回復する。
	if c.Base.Cons == 6 {
		count++
	}
	offset := 120 / count

	c.burstTickSrc = c.Core.F
	for i := 0; i < count; i++ {
		// 各氷柱はi * offsetで開始（例: 0, 40, 80 または 0, 30, 60, 90）
		// ダメージは120フレームごとに発生（前方でのみ命中するため）
		// 氷柱衍突時、半径2の範囲ダメージが発動
		// 実質的に、氷柱が1周するたびに全ターゲットが被弾
		c.Core.Tasks.Add(c.burstTickerFunc(ai, snap, c.Core.F), burstHitmark+offset*i)
	}

	c.ConsumeEnergy(51)
	if c.Base.Cons >= 6 {
		c.Core.Tasks.Add(func() { c.AddEnergy("kaeya-c6", 15) }, 52)
	}

	c.SetCDWithDelay(action.ActionBurst, 900, burstCDStart)

	// 2凸の発動カウントをリセット
	c.c2ProcCount = 0

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionJump], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}

func (c *char) burstTickerFunc(ai combat.AttackInfo, snap combat.Snapshot, src int) func() {
	return func() {
		// 元素爆発が有効か確認
		if c.Core.Status.Duration(burstKey) == 0 {
			return
		}
		// 同じ元素爆発か確認
		if c.burstTickSrc != src {
			c.Core.Log.NewEvent("kaeya burst tick ignored, src diff", glog.LogCharacterEvent, c.Index).
				Write("src", src).
				Write("new src", c.burstTickSrc)
			return
		}
		// 氷柱ダメージを発動
		c.Core.QueueAttackWithSnap(
			ai,
			snap,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 4),
			0,
		)
		// 次の氷柱Tickをキューに追加
		c.Core.Tasks.Add(c.burstTickerFunc(ai, snap, src), 120)
	}
}
