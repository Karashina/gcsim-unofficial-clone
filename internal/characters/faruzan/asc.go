package faruzan

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
)

// 固有天賦1は aimed.go と skill.go で実装:
// ファルザンが風域の創り出した「風導」状態の時、
// 狙い撃ちのチャージ時間が60%減少し、
// 結圧崩壊の渦巻きに当たった敵に秘羽の虎風の効果を付与できる。

const (
	a4Key    = "faruzan-a4"
	a4ICDKey = "faruzan-a4-icd"
)

// 「祝福の風」の効果を受けたキャラクターが通常攻撃、重撃、落下攻撃、
// 元素スキル、元素爆発で敵に風元素ダメージを与えると、
// 「ハリケーンガード」効果を得る:
// このダメージはファルザンの基礎攻撃力の32%に基づいて増加する。
// ハリケーンガードは0.8秒に1回発生可能。
// このダメージボーナスは「祝福の風」の効果終了時または
// 1回発動後に解除される。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.Element != attributes.Anemo {
			return false
		}

		switch atk.Info.AttackTag {
		case attacks.AttackTagNormal,
			attacks.AttackTagExtra,
			attacks.AttackTagPlunge,
			attacks.AttackTagElementalArt,
			attacks.AttackTagElementalArtHold,
			attacks.AttackTagElementalBurst:
		default:
			return false
		}

		active := c.Core.Player.ByIndex(atk.Info.ActorIndex)
		if active.StatusIsActive(burstBuffKey) && !c.StatusIsActive(a4ICDKey) {
			amt := 0.32 * c.Stat(attributes.BaseATK)
			if c.Core.Flags.LogDebug {
				c.Core.Log.NewEvent("faruzan a4 proc dmg add", glog.LogPreDamageMod, atk.Info.ActorIndex).
					Write("before", atk.Info.FlatDmg).
					Write("addition", amt)
			}
			atk.Info.FlatDmg += amt
			c.AddStatus(a4ICDKey, 48, false)
		}

		return false
	}, "faruzan-a4-hook")
}
