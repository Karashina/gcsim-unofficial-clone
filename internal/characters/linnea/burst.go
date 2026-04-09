package linnea

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
)

var burstFrames []int

const burstHitmark = 60

func init() {
	burstFrames = frames.InitAbilSlice(70)
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	// ルミが既にフィールドにいる場合、持続時間をリセット（形態は変更しない）
	if c.lumiActive {
		c.resetLumiDuration()
		c.Core.Log.NewEvent("Lumi duration reset by Burst (form unchanged)",
			glog.LogCharacterEvent, c.Index).
			Write("form", c.lumiForm)
	} else {
		// ルミを召喚してスーパーパワーフォームに入る
		c.summonLumi(lumiFormSuper)
		c.Core.Log.NewEvent("Linnea summons Lumi via Burst in Super Power Form",
			glog.LogCharacterEvent, c.Index)
	}

	// 初期回復
	c.QueueCharTask(func() {
		heal := burstHealFlat[c.TalentLvlBurst()] + burstHealPer[c.TalentLvlBurst()]*c.TotalDef(false)
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index,
			Target:  -1,
			Message: "Memo: Survival Guide (Initial Healing)",
			Src:     heal,
			Bonus:   c.Stat(attributes.Heal),
		})
	}, burstHitmark)

	// 継続回復を開始（12秒間、2秒間隔で6ティック）
	burstSrc := c.Core.F
	c.AddStatus(burstHealKey, burstHealDuration, true)

	for i := 1; i <= 6; i++ {
		tick := i
		delay := burstHitmark + tick*burstHealTickRate
		c.QueueCharTask(func() {
			// 回復状態がまだアクティブか確認
			if !c.StatusIsActive(burstHealKey) {
				return
			}
			_ = burstSrc // 明示的にキャプチャ（コンパイラ最適化用）

			contHeal := burstContHealFlat[c.TalentLvlBurst()] + burstContHealPer[c.TalentLvlBurst()]*c.TotalDef(false)
			c.Core.Player.Heal(info.HealInfo{
				Caller:  c.Index,
				Target:  c.Core.Player.Active(),
				Message: "Memo: Survival Guide (Continuous Healing)",
				Src:     contHeal,
				Bonus:   c.Stat(attributes.Heal),
			})
		}, delay)
	}

	// クールダウンとエネルギー消費
	c.SetCDWithDelay(action.ActionBurst, burstCD, burstHitmark)
	c.ConsumeEnergy(4)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash],
		State:           action.BurstState,
	}, nil
}
