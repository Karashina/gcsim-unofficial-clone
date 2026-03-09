package sucrose

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// スクロースが拡散反応を起こした時、パーティー内の対応する元素を持つ全キャラクター（スクロースを除く）の
// 元素熔冶が8秒間50上昇する。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}

	c.a1Buff = make([]float64, attributes.EndStatType)
	c.a1Buff[attributes.EM] = 50
	swirlfunc := func(ele attributes.Element) func(args ...interface{}) bool {
		icd := -1
		return func(args ...interface{}) bool {
			if _, ok := args[0].(*enemy.Enemy); !ok {
				return false
			}

			atk := args[1].(*combat.AttackEvent)
			if atk.Info.ActorIndex != c.Index {
				return false
			}
			// 同一フレームの場合はModを上書きしない
			if c.Core.F < icd {
				return false
			}
			icd = c.Core.F + 1

			for _, char := range c.Core.Player.Chars() {
				this := char
				if this.Base.Element != ele {
					continue
				}
				this.AddStatMod(character.StatMod{
					Base:         modifier.NewBaseWithHitlag("sucrose-a1", 480), // 8s
					AffectedStat: attributes.EM,
					Amount: func() ([]float64, bool) {
						return c.a1Buff, true
					},
				})
			}

			c.Core.Log.NewEvent("sucrose a1 triggered", glog.LogCharacterEvent, c.Index).
				Write("reaction", "swirl-"+ele.String()).
				Write("expiry", c.Core.F+480)
			return false
		}
	}

	c.Core.Events.Subscribe(event.OnSwirlCryo, swirlfunc(attributes.Cryo), "sucrose-a1-cryo")
	c.Core.Events.Subscribe(event.OnSwirlElectro, swirlfunc(attributes.Electro), "sucrose-a1-electro")
	c.Core.Events.Subscribe(event.OnSwirlHydro, swirlfunc(attributes.Hydro), "sucrose-a1-hydro")
	c.Core.Events.Subscribe(event.OnSwirlPyro, swirlfunc(attributes.Pyro), "sucrose-a1-pyro")
}

// 風霊作成・六三〇八または禁・結禅原理の雑化式で敵に命中した時、
// パーティー全員（スクロースを除く）の元素熔冶がスクロースの元素熔冶の20%分加算される。8秒間持続。
//
// - skill.go と burst.go の攻撃コールバック内で呼び出される
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}

	c.a4Buff = make([]float64, attributes.EndStatType)
	c.a4Buff[attributes.EM] = c.NonExtraStat(attributes.EM) * .20
	for i, char := range c.Core.Player.Chars() {
		if i == c.Index {
			continue // nothing for sucrose
		}
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("sucrose-a4", 480), // 8 s
			AffectedStat: attributes.EM,
			Extra:        true,
			Amount: func() ([]float64, bool) {
				return c.a4Buff, true
			},
		})
	}

	c.Core.Log.NewEvent("sucrose a4 triggered", glog.LogCharacterEvent, c.Index).
		Write("em snapshot", c.a4Buff[attributes.EM]).
		Write("expiry", c.Core.F+480)
}
