package mika

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
)

var burstFrames []int

const (
	initialHeal = 51 // pingに依存
	healKey     = "eagleplume"
	healIcdKey  = "eagleplume-icd"
)

func init() {
	burstFrames = frames.InitAbilSlice(61) // Q -> N1/Dash/Walk
	burstFrames[action.ActionSkill] = 60
	burstFrames[action.ActionJump] = 60
	burstFrames[action.ActionSwap] = 59
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	// 初期回復
	c.QueueCharTask(func() {
		heal := burstHealFirstF[c.TalentLvlBurst()] + burstHealFirstP[c.TalentLvlBurst()]*c.MaxHP()
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index,
			Target:  -1,
			Message: "Skyfeather Song",
			Src:     heal,
			Bonus:   c.Stat(attributes.Heal),
		})

		if c.Base.Cons >= 4 {
			c.c4Count = 5
		}
		c.AddStatus(healKey, 15*60, false)
	}, initialHeal)

	c.SetCD(action.ActionBurst, 18*60)
	c.ConsumeEnergy(6)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}

func (c *char) onBurstHeal() {
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		if !c.StatusIsActive(healKey) {
			return false
		}

		atk := args[1].(*combat.AttackEvent)
		if atk.Info.AttackTag != attacks.AttackTagNormal {
			return false
		}
		active := c.Core.Player.ByIndex(atk.Info.ActorIndex)
		if active.StatusIsActive(healIcdKey) {
			return false
		}
		active.AddStatus(healIcdKey, c.healIcd, true)

		heal := burstHealF[c.TalentLvlBurst()] + burstHealP[c.TalentLvlBurst()]*c.MaxHP()
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index,
			Target:  active.Index,
			Message: "Eagleplume",
			Src:     heal,
			Bonus:   c.Stat(attributes.Heal),
		})

		// ミカ自身の星翼の歌の鷲羽状態がパーティメンバーを回復した時、ミカの元素エネルギーが3回復する。
		// このエネルギー回復は、1回の星翼の歌で生成された鷲羽状態中に5回まで発動可能。
		if c.Base.Cons >= 4 && c.c4Count > 0 {
			c.AddEnergy("mika-c4", 3)
			c.c4Count--
		}

		return false
	}, "mika-eagleplume")
}
