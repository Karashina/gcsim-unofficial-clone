package heizou

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 鹿野院平蔵がフィールドに登場してから5秒間、通常攻撃速度+15%。
// また、勉ぎの心の譲り重ねスタックを1層獲得する。この効果は10秒に1回発動可能。
func (c *char) c1() {
	const c1Icd = "heizou-c1-icd"
	// ステータスモッドがデバッグビューに表示されるため、ログ値は保存しない
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(args ...interface{}) bool {
		if c.StatusIsActive(c1Icd) {
			return false
		}
		next := args[1].(int)
		if next != c.Index {
			return false
		}
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("heizou-c1", 300), // 5s
			AffectedStat: attributes.AtkSpd,
			Amount: func() ([]float64, bool) {
				return c.c1buff, true
			},
		})
		c.addDecStack()
		c.AddStatus(c1Icd, 600, true)
		return false
	}, "heizou enter")
}

// 彩風の暴風の最初の「風の患い」爆発は平蔵の元素エネルギーを9回復する。
// その後の爆発はそれぞれ追加で1.5エネルギー回復。
// 1回の彩風の暴風で合訓13.5エネルギー回復可能。
func (c *char) c4(i int) {
	switch i {
	case 1:
		c.AddEnergy("heizou c4", 9.0)
	case 2, 3, 4:
		c.AddEnergy("heizou c4", 1.5)
	}
}

// 譲り重ねスタック1層あたり、勉ぎの心の会心率+4%。
// 平蔵が「釈然」状態で発動した場合、会心ダメージ+32%。
func (c *char) c6() (float64, float64) {
	cr := 0.04 * float64(c.decStack)

	cd := 0.0
	if c.decStack == 4 {
		cd = 0.32
	}

	if cr > 0 {
		c.Core.Log.NewEvent("heizou-c6 adding stats", glog.LogCharacterEvent, c.Index).
			Write("cr", cr).
			Write("cd", cd)
	}

	return cr, cd
}
