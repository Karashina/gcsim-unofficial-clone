package linnea

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
)

var burstFrames []int

const burstHitmark = 96 // Q -> initial heal: 96f

func init() {
	burstFrames = frames.InitAbilSlice(70)
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	// if Lumi is already active, reset duration (form unchanged)
	if c.lumiActive {
		c.resetLumiDuration()
		c.Core.Log.NewEvent("Lumi duration reset by Burst (form unchanged)",
			glog.LogCharacterEvent, c.Index).
			Write("form", c.lumiForm)
	} else {
		// summon Lumi and enter Super Power Form
		c.summonLumi(lumiFormSuper, lumiFirstTickFromQ)
		c.Core.Log.NewEvent("Linnea summons Lumi via Burst in Super Power Form",
			glog.LogCharacterEvent, c.Index)
	}

	// initial heal (Q -> heal: 96f)
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

	// start continuous healing (Q -> cont. heal start: 158f, interval: 60f, ticks: 12)
	burstSrc := c.Core.F
	c.AddStatus(burstHealKey, burstHealDuration, true)

	for i := 0; i < burstHealTicks; i++ {
		tick := i
		delay := burstContHealStart + tick*burstHealTickRate
		c.QueueCharTask(func() {
			// check if heal status is still active
			if !c.StatusIsActive(burstHealKey) {
				return
			}
			_ = burstSrc // explicit capture for compiler optimization

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

	// cooldown (CD start: 2f) and energy consumption (4f)
	c.SetCDWithDelay(action.ActionBurst, burstCD, burstCDDelay)
	c.ConsumeEnergy(4)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash],
		State:           action.BurstState,
	}, nil
}
