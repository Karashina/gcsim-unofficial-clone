package kaeya

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/shield"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 1凸:
// 氷元素の影響を受けた敵に対するカイアの通常攻撃と重撃の会心率+15%。
func (c *char) c1() {
	m := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("kaeya-c1", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			e, ok := t.(*enemy.Enemy)
			if !ok {
				return nil, false
			}
			if atk.Info.AttackTag != attacks.AttackTagNormal && atk.Info.AttackTag != attacks.AttackTagExtra {
				return nil, false
			}
			if !e.AuraContains(attributes.Cryo, attributes.Frozen) {
				return nil, false
			}
			m[attributes.CR] = 0.15
			return m, true
		},
	})
}

// 2凸:
// 冰雪の輪の持続中に敵を倒すたびに持続時間が2.5秒延長、最大15秒まで。
func (c *char) c2() {
	c.Core.Events.Subscribe(event.OnTargetDied, func(args ...interface{}) bool {
		_, ok := args[0].(*enemy.Enemy)
		// 敵でなければ無視
		if !ok {
			return false
		}
		// 元素爆発が有効でない場合は無視
		if c.Core.Status.Duration(burstKey) == 0 {
			return false
		}
		// 延長上限に達した場合は無視
		if c.c2ProcCount > 2 {
			return false
		}
		// 元素爆発持続時間の延長段階
		// 8秒
		// 10.5秒 (前回から+2.5秒)
		// 13秒 (前回から+2.5秒)
		// 15秒 (前回から+2.0秒、上限15秒のため)
		extension := 150
		if c.c2ProcCount == 2 {
			extension = 120
		}
		c.Core.Status.Extend(burstKey, extension)
		c.c2ProcCount++
		c.Core.Log.NewEvent("kaeya-c2 proc'd", glog.LogCharacterEvent, c.Index).
			Write("c2ProcCount", c.c2ProcCount).
			Write("extension", extension)
		return false
	}, "kaeya-c2")
}

// 4凸:
// カイアのHPが20%を下回ると自動発動:
// カイアの最大HPの30%のダメージを吸収するシールドを生成。20秒持続。
// 氷元素ダメージに対して250%の吸収効率。
// 60秒に1回のみ発動可能。
func (c *char) c4() {
	c.Core.Events.Subscribe(event.OnPlayerHPDrain, func(args ...interface{}) bool {
		di := args[0].(*info.DrainInfo)
		if di.Amount <= 0 {
			return false
		}
		if c.Core.F < c.c4icd && c.c4icd != 0 {
			return false
		}
		maxhp := c.MaxHP()
		if c.CurrentHPRatio() < 0.2 {
			c.c4icd = c.Core.F + 3600
			c.Core.Player.Shields.Add(&shield.Tmpl{
				ActorIndex: c.Index,
				Target:     c.Index,
				Src:        c.Core.F,
				ShieldType: shield.KaeyaC4,
				Name:       "Kaeya C4",
				HP:         .3 * maxhp,
				Ele:        attributes.Cryo,
				Expires:    c.Core.F + 1200,
			})
		}
		return false
	}, "kaeya-c4")
}
