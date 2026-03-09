package aino

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
)

var burstFrames []int

const (
	burstDuration       = 14 * 60 // 14秒
	burstTickInterval   = 90      // 90フレーム（デフォルト）
	burstTickIntervalA1 = 42      // 42フレーム（ムーンサインが昇順の時）
	burstKey            = "aino-burst"
)

func init() {
	burstFrames = frames.InitAbilSlice(62)
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	c.burstSrc = c.Core.F

	// 固有天賦1に基づき間隔を決定
	interval := burstTickInterval
	radius := 3.0
	if c.MoonsignAscendant {
		interval = burstTickIntervalA1
		radius = 5.0
	}

	c.QueueCharTask(func() {
		// 水球攻撃をスケジュール
		numTicks := burstDuration / interval
		for i := 0; i < numTicks; i++ {
			c.Core.Tasks.Add(c.burstTick(c.burstSrc, radius), interval*i)
		}
	}, 120)

	c.AddStatus(burstKey, burstDuration, false)

	c.SetCD(action.ActionBurst, 13.5*60) // CD 13.5秒
	c.ConsumeEnergy(7)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash],
		State:           action.BurstState,
	}, nil
}

func (c *char) burstTick(src int, radius float64) func() {
	return func() {
		if src != c.burstSrc {
			return
		}
		if !c.StatusIsActive(burstKey) {
			return
		}

		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Precision Hydronic Cooler (Water Ball)",
			AttackTag:  attacks.AttackTagElementalBurst,
			ICDTag:     attacks.ICDTagElementalBurst,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Hydro,
			Durability: 25,
			Mult:       burst[c.TalentLvlBurst()],
		}
		if c.MoonsignAscendant {
			ai.ICDGroup = attacks.ICDGroupAinoBurstEnhanced
		}
		if c.Base.Ascension >= 4 {
			ai.FlatDmg += 0.5 * c.Stat(attributes.EM)
		}

		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, radius),
			0,
			0,
		)
	}
}
