package xiao

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
)

var burstFrames []int

const (
	burstStart   = 57
	burstBuffKey = "xiaoburst"
)

func init() {
	burstFrames = frames.InitAbilSlice(82) // Q -> N1/E
	burstFrames[action.ActionDash] = 59    // Q -> D
	burstFrames[action.ActionJump] = 60    // Q -> J
	burstFrames[action.ActionSwap] = 66    // Q -> Swap
}

// 魉の元素爆発ダメージ状態を設定
func (c *char) Burst(p map[string]int) (action.Info, error) {
	var hpICD int
	hpICD = 0

	// 以前のコードによれば、元素爆発の持続時間はアニメーション完了後からカウント開始
	// TODO: ライブラリにその記載はない
	c.AddStatus(burstBuffKey, 900+burstStart, true)
	c.qStarted = c.Core.F
	c.a1()

	// HPドレイン - 元素爆発発動後、1秒ごとにHPが減少
	// ゲームプレイ動画によると、HPティックはアニメーション完了後に開始
	for i := burstStart + 60; i < 900+burstStart; i++ {
		c.Core.Tasks.Add(func() {
			if c.StatusIsActive(burstBuffKey) && c.Core.F >= hpICD {
				// TODO: これがヒットラグの影響を受けるか不明
				hpICD = c.Core.F + 60
				c.Core.Player.Drain(info.DrainInfo{
					ActorIndex: c.Index,
					Abil:       "Bane of All Evil",
					Amount:     burstDrain[c.TalentLvlBurst()] * c.CurrentHP(),
				})
			}
		}, i)
	}

	c.SetCDWithDelay(action.ActionBurst, 18*60, 29)
	c.ConsumeEnergy(36)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}

// 魉がフィールドを離れたら元素爆発を早期終了するフック
func (c *char) onExitField() {
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(_ ...interface{}) bool {
		c.DeleteStatus(burstBuffKey)
		c.DeleteStatus(a1Key)
		return false
	}, "xiao-exit")
}
