package dendro

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
)

var burstFrames [][]int

const (
	burstKey       = "travelerdendro-q"
	burstHitmark   = 91
	leaLotusAppear = 54
)

func init() {
	burstFrames = make([][]int, 2)

	// 男性
	burstFrames[0] = frames.InitAbilSlice(58)
	burstFrames[0][action.ActionSwap] = 57 // Q -> Swap

	// 女性
	burstFrames[1] = frames.InitAbilSlice(58)
	burstFrames[1][action.ActionSwap] = 57 // Q -> Swap
}

func (c *Traveler) Burst(p map[string]int) (action.Info, error) {
	c.SetCD(action.ActionBurst, 1200)
	c.ConsumeEnergy(2)

	// 持続時間は最初のヒットマークからカウント

	c.Core.Tasks.Add(func() {
		s := c.newLeaLotusLamp()

		if c.Base.Ascension >= 1 {
			// 固有天賦1は毎秒スタックを追加
			for delay := 0; delay <= s.Gadget.Duration; delay += 60 {
				c.a1Stack(delay)
			}
			// 固有天賦1/6凸のバフは0.3秒ごとにティックし、1秒間適用。おそらくガジェット出現からカウント - Kolibri
			for delay := 0; delay <= s.Gadget.Duration; delay += 0.3 * 60 {
				c.a1Buff(delay)
			}
		}

		if c.Base.Cons >= 6 {
			for delay := 0; delay <= s.Gadget.Duration; delay += 0.3 * 60 {
				c.c6Buff(delay)
			}
		}
		c.Core.Combat.AddGadget(s)
	}, leaLotusAppear)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames[c.gender]),
		AnimationLength: burstFrames[c.gender][action.InvalidAction],
		CanQueueAfter:   burstFrames[c.gender][action.ActionSwap], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}
