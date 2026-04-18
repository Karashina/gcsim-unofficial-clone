package linnea

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
)

var burstFrames []int

const burstHitmark = 96 // Q->回復: 96f (初期回復ヒットマーク)

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
		c.summonLumi(lumiFormSuper, lumiFirstTickFromQ)
		c.Core.Log.NewEvent("Linnea summons Lumi via Burst in Super Power Form",
			glog.LogCharacterEvent, c.Index)
	}

	// 初期回復 (Q->回復: 96f)
	c.QueueCharTask(func() {
		heal := burstHealFlat[c.TalentLvlBurst()] + burstHealPer[c.TalentLvlBurst()]*c.TotalDef(false)
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index,
			Target:  -1,
			Message: "Memo: Survival Guide (Initial Healing)",
			Src:     heal,
			Bonus:   c.Stat(attributes.Heal),
		})
	}, burstInitHealDelay)

	// 継続回復を開始 (Q->継続回復開始: 158f, 回復間隔: 60f, 回復回数: 12回)
	burstSrc := c.Core.F
	c.AddStatus(burstHealKey, burstHealDuration, true)

	for i := 0; i < burstHealTicks; i++ {
		tick := i
		delay := burstContHealStart + tick*burstHealTickRate
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

	// クールダウン (CT開始: 2f) とエネルギー消費 (4f)
	c.SetCDWithDelay(action.ActionBurst, burstCD, burstCDDelay)
	c.ConsumeEnergy(4)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash],
		State:           action.BurstState,
	}, nil
}
