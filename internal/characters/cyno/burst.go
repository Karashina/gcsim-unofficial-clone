package cyno

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var burstFrames []int

const (
	burstKey = "cyno-q"
)

func init() {
	burstFrames = frames.InitAbilSlice(86) // Q -> J
	burstFrames[action.ActionAttack] = 84
	burstFrames[action.ActionSkill] = 84
	burstFrames[action.ActionDash] = 84
	burstFrames[action.ActionSwap] = 83
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	c.burstExtension = 0 // 元素爆発ごとに延長可能回数をリセット
	c.c4Counter = 0      // 第4命ノ星座スタックをリセット
	c.c6Stacks = 0       // 同上

	m := make([]float64, attributes.EndStatType)
	m[attributes.EM] = 100
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag(burstKey, 712), // 112f追加持続
		AffectedStat: attributes.EM,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})
	c.burstSrc = c.Core.F
	src := c.Core.F
	// セノが元素爆発を延長した場合、元素スキルCDを正しく設定する必要がある
	c.QueueCharTask(func() { c.onBurstExpiry(src) }, 713+240)
	c.QueueCharTask(func() { c.onBurstExpiry(src) }, 713+480)

	if c.Base.Ascension >= 1 {
		c.QueueCharTask(c.a1, 328)
	}
	c.SetCD(action.ActionBurst, 1200)
	c.ConsumeEnergy(3)

	if c.Base.Cons >= 1 {
		c.c1()
	}
	c.c6Init()

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}

func (c *char) tryBurstPPSlide(hitmark int) {
	duration := c.StatusDuration(burstKey)
	if 0 < duration && duration < hitmark {
		c.ExtendStatus(burstKey, hitmark-duration+1)
		c.Core.Log.NewEvent("pp slide activated", glog.LogCharacterEvent, c.Index).
			Write("expiry", c.StatusExpiry(burstKey))
		src := c.burstSrc
		c.QueueCharTask(func() {
			c.onBurstExpiry(src)
		}, hitmark-duration+3) // 3f（元素爆発は2fで終了するため）
	}
}

func (c *char) onExitField() {
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(args ...interface{}) bool {
		if !c.StatusIsActive(burstKey) {
			return false
		}
		prev := args[0].(int)
		if prev == c.Index {
			c.DeleteStatus(burstKey)
			c.onBurstExpiry(c.burstSrc)
		}
		return false
	}, "cyno-burst-clear")
}

func (c *char) onBurstExpiry(burstSrc int) {
	if burstSrc != c.burstSrc {
		return
	}
	if c.StatusIsActive(burstKey) {
		return
	}
	c.burstSrc = -1 // 他の元素爆発関数を呼ばないようにする
}
