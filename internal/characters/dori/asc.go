package dori

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
)

// ジニーに接続されたキャラクターが感電・超伝導・過負荷・激化・激化（超）・超開花・
// 雷元素の拡散・結晶反応をトリガーすると、魔除の灯のCDが1秒短縮される。
// この効果は3秒に1回発動可能。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}

	const icdKey = "dori-a1"
	icd := 180 // 3s * 60
	//nolint:unparam // ignoring for now, event refactor should get rid of bool return of event sub
	reduce := func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)

		if c.Core.Player.Active() != atk.Info.ActorIndex { // フィールド上のキャラクターのみ
			return false
		}
		if c.StatusIsActive(icdKey) {
			return false
		}
		c.AddStatus(icdKey, icd, true)
		c.ReduceActionCooldown(action.ActionSkill, 60)
		c.Core.Log.NewEvent("dori a1 proc", glog.LogCharacterEvent, c.Index).
			Write("reaction", atk.Info.Abil).
			Write("new cd", c.Cooldown(action.ActionSkill))
		return false
	}

	reduceNoGadget := func(args ...interface{}) bool {
		if _, ok := args[0].(*enemy.Enemy); !ok {
			return false
		}
		return reduce(args...)
	}

	c.Core.Events.Subscribe(event.OnOverload, reduceNoGadget, "dori-a1")
	c.Core.Events.Subscribe(event.OnElectroCharged, reduceNoGadget, "dori-a1")
	c.Core.Events.Subscribe(event.OnSuperconduct, reduceNoGadget, "dori-a1")
	c.Core.Events.Subscribe(event.OnQuicken, reduceNoGadget, "dori-a1")
	c.Core.Events.Subscribe(event.OnAggravate, reduceNoGadget, "dori-a1")
	c.Core.Events.Subscribe(event.OnHyperbloom, reduce, "dori-a1")
	c.Core.Events.Subscribe(event.OnCrystallizeElectro, reduceNoGadget, "dori-a1")
	c.Core.Events.Subscribe(event.OnSwirlElectro, reduceNoGadget, "dori-a1")
}

// トラブルシューターショットまたはアフターサービス弾が敵に命中すると、
// ドリーは元素チャージ効率100%ごとに元素エネルギーを5回復する。
// 魔除の灯1回につきエネルギー回復は1回のみ発動可能で、
// 最大15まで回復可能。
func (c *char) makeA4CB() combat.AttackCBFunc {
	if c.Base.Ascension < 4 {
		return nil
	}

	done := false
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if done {
			return
		}
		done = true

		a4Energy := a.AttackEvent.Snapshot.Stats[attributes.ER] * 5
		if a4Energy > 15 {
			a4Energy = 15
		}
		c.AddEnergy("dori-a4", a4Energy)
	}
}
