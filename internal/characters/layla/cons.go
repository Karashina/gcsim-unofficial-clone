package layla

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const c4Key = "layla-c4"

// ひとしきりの夜がシューティングスターを発射し始めると、近くのパーティメンバー全員に「暁の星」効果を付与し、
// レイラの最大HPの5%に基づき通常攻撃・重撃ダメージが増加する。
// 暁の星は最大3秒間持続し、通常攻撃または重撃ダメージを与えた0.05秒後に削除される。
func (c *char) c4() {
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)
		if ae.Info.AttackTag != attacks.AttackTagNormal && ae.Info.AttackTag != attacks.AttackTagExtra {
			return false
		}

		char := c.Core.Player.ByIndex(ae.Info.ActorIndex)
		if !char.StatusIsActive(c4Key) {
			return false
		}

		dmgAdded := 0.05 * c.MaxHP()
		ae.Info.FlatDmg += dmgAdded

		c.QueueCharTask(func() { char.DeleteStatus(c4Key) }, 0.05*60)

		c.Core.Log.NewEvent("layla c4 adding damage", glog.LogPreDamageMod, ae.Info.ActorIndex).
			Write("damage_added", dmgAdded)

		return false
	}, "layla-c4")
}

// ひとしきりの夜のシューティングスターのダメージが40%増加し、星竜の蜜恋のStarlight Slugのダメージが40%増加する。
// また、ひとしきりの夜による星の生成間隔が20%短縮される。
func (c *char) c6() {
	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = 0.4

	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("layla-c6", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagElementalBurst && atk.Info.Abil != "Shooting Star" {
				return nil, false
			}
			return m, true
		},
	})
}
