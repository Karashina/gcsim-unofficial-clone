package kazuha

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 万葉の狂風発動時に水/炎/氷/雷元素に接触すると、その元素を吸収し、
// 流乱武花落下攻撃が効果終了前に使用されると、
// 吸収した元素の攻撃力200%の追加ダメージを与える。これは落下攻撃ダメージとみなされる。
// 元素吸収は万葉の狂風の使用ごとに1回のみ発生可能。
//
// - スキル発動失敗を避けるため、突破レベルチェックは skill.go で実施
func (c *char) absorbCheckA1(src, count, maxcount int) func() {
	return func() {
		if count == maxcount {
			return
		}
		c.a1Absorb = c.Core.Combat.AbsorbCheck(c.Index, c.a1AbsorbCheckLocation, attributes.Pyro, attributes.Hydro, attributes.Electro, attributes.Cryo)

		if c.a1Absorb != attributes.NoElement {
			c.Core.Log.NewEventBuildMsg(glog.LogCharacterEvent, c.Index,
				"kazuha a1 absorbed ", c.a1Absorb.String(),
			)
			return
		}
		// それ以外はキューに追加
		c.Core.Tasks.Add(c.absorbCheckA1(src, count+1, maxcount), 6)
	}
}

// 拡散反応トリガー時、吸収した元素の元素ダメージボーナス+0.04%を
// 全パーティメンバーに8秒間付与する（元素熟知1あたり）。
// 異なる元素のボーナスは共存可能。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}

	m := make([]float64, attributes.EndStatType)

	swirlfunc := func(ele attributes.Stat, key string) func(args ...interface{}) bool {
		icd := -1
		return func(args ...interface{}) bool {
			if _, ok := args[0].(*enemy.Enemy); !ok {
				return false
			}

			atk := args[1].(*combat.AttackEvent)
			if atk.Info.ActorIndex != c.Index {
				return false
			}
			// 同一フレームではモディファイアーを上書きしない
			if c.Core.F < icd {
				return false
			}
			icd = c.Core.F + 1

			// 元素熟知を再計算
			dmg := 0.0004 * c.NonExtraStat(attributes.EM)

			for _, char := range c.Core.Player.Chars() {
				char.AddStatMod(character.StatMod{
					Base:         modifier.NewBaseWithHitlag("kazuha-a4-"+key, 60*8),
					AffectedStat: ele,
					Extra:        true,
					Amount: func() ([]float64, bool) {
						clear(m)
						m[ele] = dmg
						return m, true
					},
				})
			}

			c.Core.Log.NewEvent("kazuha a4 proc", glog.LogCharacterEvent, c.Index).
				Write("reaction", ele.String())

			return false
		}
	}

	c.Core.Events.Subscribe(event.OnSwirlCryo, swirlfunc(attributes.CryoP, "cryo"), "kazuha-a4-cryo")
	c.Core.Events.Subscribe(event.OnSwirlElectro, swirlfunc(attributes.ElectroP, "electro"), "kazuha-a4-electro")
	c.Core.Events.Subscribe(event.OnSwirlHydro, swirlfunc(attributes.HydroP, "hydro"), "kazuha-a4-hydro")
	c.Core.Events.Subscribe(event.OnSwirlPyro, swirlfunc(attributes.PyroP, "pyro"), "kazuha-a4-pyro")
}
