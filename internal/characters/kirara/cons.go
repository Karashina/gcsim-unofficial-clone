package kirara

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/shield"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	c4IcdStatus = "kirara-c4-icd"
	c6Status    = "kirara-c6"
)

// 2凸は未実装、マルチ専用
// キララが緋球玄縄のUrgent Neko Parcel状態の時、衝突した他のパーティメンバーに緊急輸送シールドを付与する。
// 緊急輸送シールドのダメージ吸収量は緋球玄縄の安全輸送シールドの最大吸収量の40%で、
// 草元素ダメージを250%の効率で吸収する。
// 緊急輸送シールドは12秒間持続し、各キャラクターに対して10秒に1回発動可能。

// 安全輸送シールドまたは緊急輸送シールドで保護されたアクティブキャラの通常攻撃・重撃・落下攻撃が敵に命中すると、
// キララが小型Cat Grass Cardamomで協力攻撃を行い、攻撃力200%の草元素ダメージを与える。
// このダメージは元素爆発ダメージとみなされる。この効果は3.8秒に1回発動可能。このCDはパーティ全員で共有。
func (c *char) c4() {
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		if c.StatusIsActive(c4IcdStatus) {
			return false
		}
		existingShield := c.Core.Player.Shields.Get(shield.KiraraSkill)
		if existingShield == nil {
			return false
		}

		atk := args[1].(*combat.AttackEvent)
		switch atk.Info.AttackTag {
		case attacks.AttackTagNormal,
			attacks.AttackTagExtra,
			attacks.AttackTagPlunge:
		default:
			return false
		}
		t := args[0].(combat.Target)

		// TODO: スナップショット？ダメージ遅延？
		ai := combat.AttackInfo{
			ActorIndex:         c.Index,
			Abil:               "Steed of Skanda",
			AttackTag:          attacks.AttackTagElementalBurst,
			ICDTag:             attacks.ICDTagElementalBurst,
			ICDGroup:           attacks.ICDGroupDefault,
			StrikeType:         attacks.StrikeTypeDefault,
			Element:            attributes.Dendro,
			Durability:         25,
			Mult:               2,
			CanBeDefenseHalted: true,
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(t, nil, 2), 0, 0)
		c.AddStatus(c4IcdStatus, 3.8*60, true)
		return false
	}, "kirara-c4")
}

// キララが元素スキルまたは元素爆発を使用後15秒間、近くのパーティ全員の全元素ダメージ+12%。
func (c *char) c6() {
	for _, char := range c.Core.Player.Chars() {
		char.AddStatMod(character.StatMod{
			Base: modifier.NewBaseWithHitlag(c6Status, 15*60),
			Amount: func() ([]float64, bool) {
				return c.c6Buff, true
			},
		})
	}
}
