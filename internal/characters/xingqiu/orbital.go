package xingqiu

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
)

// 新しいオービタルを開始または既にアクティブなら延長する。durationは持続時間、
// delayは最初のTick開始までの遅延
func (c *char) applyOrbital(duration, delay int) {
	src := c.Core.F
	c.Core.Log.NewEvent(
		"Applying orbital", glog.LogCharacterEvent, c.Index,
	).Write(
		"current status", c.StatusExpiry(orbitalKey),
	)
	// オービタルが既にアクティブか確認。アクティブなら持続時間を延長
	// そうでなければ最初のTickを開始
	if !c.orbitalActive {
		// ヒットラグ影響キューを使用
		c.QueueCharTask(c.orbitalTickTask(src), delay)
		c.orbitalActive = true
		c.Core.Log.NewEvent(
			"orbital applied", glog.LogCharacterEvent, c.Index,
		).Write(
			"expected end", src+900,
		).Write(
			"next expected tick", src+40,
		)
	}
	c.AddStatus(orbitalKey, duration, true)
	c.Core.Log.NewEvent(
		"orbital duration extended", glog.LogCharacterEvent, c.Index,
	).Write(
		"new expiry", c.StatusExpiry(orbitalKey),
	)
}

func (c *char) orbitalTickTask(src int) func() {
	return func() {
		c.Core.Log.NewEvent(
			"orbital checking tick", glog.LogCharacterEvent, c.Index,
		).Write(
			"expiry", c.StatusExpiry(orbitalKey),
		).Write(
			"src", src,
		)
		if !c.StatusIsActive(orbitalKey) {
			c.orbitalActive = false
			return
		}

		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Xingqiu Orbital",
			AttackTag:  attacks.AttackTagNone,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Hydro,
			Durability: 25,
		}
		c.Core.Log.NewEvent(
			"orbital ticked", glog.LogCharacterEvent, c.Index,
		).Write(
			"next expected tick", c.Core.F+135,
		).Write(
			"expiry", c.StatusExpiry(orbitalKey),
		).Write(
			"src", src,
		)

		// 次のインスタンスをキューに追加
		c.QueueCharTask(c.orbitalTickTask(src), 135)

		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 1.2), -1, 1)
	}
}
