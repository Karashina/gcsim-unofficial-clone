package candace

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const c6ICDKey = "candace-c6-icd"

// 聖儀・蒼鷺の庇護が敵に命中した時、
// キャンディスのHP上限が15秒間20%増加する。
func (c *char) c2() {
	m := make([]float64, attributes.EndStatType)
	m[attributes.HPP] = 0.2
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag("candace-c2", 15*60),
		AffectedStat: attributes.HPP,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})
}

// 聖儀・灰鶺鴒の潮によって赤冠の祈りの影響を受けたキャラクター（キャンディス自身を除く）が
// 通常攻撃で敵に元素ダメージを与えた時、キャンディスのHP上限15%に相当する
// 水元素範囲ダメージを与える攻撃波が発生する。この効果は2.3秒ごとに1回発動可能で、
// 元素爆発ダメージとみなされる。
func (c *char) c6() {
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		dmg := args[2].(float64)
		if atk.Info.AttackTag != attacks.AttackTagNormal {
			return false
		}
		if atk.Info.Element == attributes.Physical || atk.Info.Element == attributes.NoElement {
			return false
		}
		if atk.Info.ActorIndex != c.Core.Player.Active() {
			return false
		}
		if atk.Info.ActorIndex == c.Index {
			return false
		}
		if !c.StatusIsActive(burstKey) {
			return false
		}
		if c.StatusIsActive(c6ICDKey) {
			return false
		}
		if dmg == 0 {
			return false
		}
		c.AddStatus(c6ICDKey, 138, true)
		ai := combat.AttackInfo{
			ActorIndex:         c.Index,
			Abil:               "The Overflow (C6)",
			AttackTag:          attacks.AttackTagElementalBurst,
			ICDTag:             attacks.ICDTagNone,
			ICDGroup:           attacks.ICDGroupDefault,
			StrikeType:         attacks.StrikeTypeDefault,
			Element:            attributes.Hydro,
			Durability:         25,
			FlatDmg:            0.15 * c.MaxHP(),
			CanBeDefenseHalted: true,
		}
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 3.5),
			waveHitmark,
			waveHitmark,
		)
		return false
	}, "candace-c6")
}
