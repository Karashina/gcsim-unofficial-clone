package ayato

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
)

// 神里流・鏡花には以下の特性がある:
//
// - 使用後、神里綾人は波闃のスタックを2獲得する。
func (c *char) a1OnSkill() {
	if c.Base.Ascension < 1 {
		return
	}
	c.stacks = 2
	c.Core.Log.NewEvent("ayato a1 proc'd", glog.LogCharacterEvent, c.Index)
}

// 神里流・鏡花には以下の特性がある:
//
// - 水分身が爆発すると、綾人は最大スタック数に等しい波闃効果を獲得する。
func (c *char) a1OnExplosion() {
	if c.Base.Ascension < 1 {
		return
	}
	c.stacks = c.stacksMax
	c.Core.Log.NewEvent("ayato a1 set namisen stacks to max", glog.LogCharacterEvent, c.Index).
		Write("stacks", c.stacks)
}

// 神里綾人がフィールド上にいないかつ元素エネルギーが40未満の場合、毎秒元素エネルギーを2回復する。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	if c.Core.Player.Active() == c.Index {
		return
	}
	if c.Energy >= 40 {
		return
	}
	c.AddEnergy("ayato-a4", 2)
	c.Core.Tasks.Add(c.a4, 60)
}
