package tighnari

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// Tighnariの重撃の会心率が15%上昇する。
func (c *char) c1() {
	m := make([]float64, attributes.EndStatType)
	m[attributes.CR] = 0.15
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("tighnari-c1", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagExtra {
				return nil, false
			}
			return m, true
		},
	})
}

// 識罪縄用素の草雷フィールド内に敵がいる時、Tighnariの草元素ダメージが20%上昇する。
// フィールドの持続時間が終了するか、フィールド内に敵がいなくなった場合、効果は最大6秒間持続する。
func (c *char) c2() {
	m := make([]float64, attributes.EndStatType)
	m[attributes.DendroP] = .2
	for i := 0; i < 8*60; i += 30 {
		c.Core.Tasks.Add(func() {
			if !c.Core.Combat.Player().IsWithinArea(c.skillArea) {
				return
			}
			c.AddStatMod(character.StatMod{
				Base:         modifier.NewBase("tighnari-c2", 6*60),
				AffectedStat: attributes.DendroP,
				Amount: func() ([]float64, bool) {
					return m, true
				},
			})
		}, i)
	}
}

// 織用素翠の祠矢が放たれた時、近くのパーティメンバー全員の元素熔通が8秒間60上昇する。
// TODO: 織用素翠の祠矢が燃焼・開花・激化・拡散反応を起こした場合、
// 元素熔通がさらに60上昇する。この場合、バフの持続時間も更新される。
func (c *char) c4() {
	c.Core.Events.Subscribe(event.OnBurst, func(args ...interface{}) bool {
		if c.Core.Player.Active() != c.Index {
			return false
		}

		m := make([]float64, attributes.EndStatType)
		m[attributes.EM] = 60
		for _, char := range c.Core.Player.Chars() {
			char.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag("tighnari-c4", 8*60),
				AffectedStat: attributes.EM,
				Amount: func() ([]float64, bool) {
					return m, true
				},
			})
		}

		return false
	}, "tighnari-c4")

	f := func(args ...interface{}) bool {
		if _, ok := args[0].(*enemy.Enemy); !ok {
			return false
		}

		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != c.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagElementalBurst {
			return false
		}

		m := make([]float64, attributes.EndStatType)
		m[attributes.EM] = 120
		for _, char := range c.Core.Player.Chars() {
			char.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag("tighnari-c4", 8*60),
				AffectedStat: attributes.EM,
				Amount: func() ([]float64, bool) {
					return m, true
				},
			})
		}

		return false
	}
	c.Core.Events.Subscribe(event.OnBurning, f, "tighnari-c4-burning")
	c.Core.Events.Subscribe(event.OnBloom, f, "tighnari-c4-bloom")
	c.Core.Events.Subscribe(event.OnQuicken, f, "tighnari-c4-quicken")
	c.Core.Events.Subscribe(event.OnSpread, f, "tighnari-c4-spread")
}
