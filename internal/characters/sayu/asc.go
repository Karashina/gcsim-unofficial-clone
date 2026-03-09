package sayu

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
)

const a1ICDKey = "sayu-a1-icd"

// 早柚が拡散反応を発動した時、全キャラクターと近くの味方を300HP回復する。
// さらに元素熔研、1ポイントごとに1.2HP追加回復する。
// この効果は2秒に1回発動可能。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	swirlfunc := func(args ...interface{}) bool {
		if _, ok := args[0].(*enemy.Enemy); !ok {
			return false
		}

		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != c.Index {
			return false
		}
		if c.Core.Player.Active() != c.Index {
			return false
		}
		if c.StatusIsActive(a1ICDKey) {
			return false
		}
		c.AddStatus(a1ICDKey, 120, true) // 2s

		if c.Base.Cons >= 4 {
			c.AddEnergy("sayu-c4", 1.2)
		}

		heal := 300 + c.Stat(attributes.EM)*1.2
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index,
			Target:  -1,
			Message: "Someone More Capable",
			Src:     heal,
			Bonus:   c.Stat(attributes.Heal),
		})

		return false
	}

	c.Core.Events.Subscribe(event.OnSwirlCryo, swirlfunc, "sayu-a1-cryo")
	c.Core.Events.Subscribe(event.OnSwirlElectro, swirlfunc, "sayu-a1-electro")
	c.Core.Events.Subscribe(event.OnSwirlHydro, swirlfunc, "sayu-a1-hydro")
	c.Core.Events.Subscribe(event.OnSwirlPyro, swirlfunc, "sayu-a1-pyro")
}

// よぶぶきの術・ムジナフルリーが生成したムジムジだるまが以下の効果を得る:
//
// - キャラクターを回復する際、回復対象の近くのキャラクターも回復量の20%分回復する
//   - コープでのみ関連
//
// - 敵への攻撃範囲が拡大される
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	c.qTickRadius = 3.5
}
