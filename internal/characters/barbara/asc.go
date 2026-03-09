package barbara

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
)

// 公演、開幕♪のメロディーループ内のキャラクターのスタミナ消費が12%減少する。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	// 固有天賦1はバーバラのスキル持続時間（900フレーム）の間持続
	c.Core.Player.AddStamPercentMod("barb-a1-stam", skillDuration, func(a action.Action) (float64, bool) {
		return -0.12, false
	})
}

// アクティブキャラクターが元素オーブ/粒子を取得した時、公演、開幕♪のメロディーループの持続時間が1秒延長される。
// 最大延長は5秒。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	c.Core.Events.Subscribe(event.OnParticleReceived, func(_ ...interface{}) bool {
		if c.Core.Status.Duration(barbSkillKey) == 0 {
			return false
		}
		if c.a4extendCount == 5 {
			return false
		}

		c.a4extendCount++
		c.Core.Status.Extend(barbSkillKey, 60)

		c.Core.Log.NewEvent("barbara skill extended from a4", glog.LogCharacterEvent, c.Index)

		return false
	}, "barbara-a4")
}
