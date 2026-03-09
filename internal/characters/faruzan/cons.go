package faruzan

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 4凸: 風域の渦巻きが敵に命中した数に応じてエネルギーを回復:
// 1体の敵に命中するとエネルギー2回復。
// 追加の敵ごとにさらに0.5回復。
// 1回の渦巻きで最大4エネルギーまで回復可能。
func (c *char) makeC4Callback() func(combat.AttackCB) {
	if c.Base.Cons < 4 {
		return nil
	}
	count := 0
	return func(a combat.AttackCB) {
		if count > 4 {
			return
		}
		amt := 0.5
		if count == 0 {
			amt = 2
		}
		count++
		c.AddEnergy("faruzan-c4", amt)
	}
}

// 6凸: 「祝福の風」の効果を受けたキャラクターが風元素ダメージを与えると
// 会心ダメージ+40%。「祝福の風」の効果中にアクティブキャラが
// ダメージを与えると、ハリケーンアローを追加発射する。
// この効果は2.5秒に1回トリガー可能。
func (c *char) c6Buff(char *character.CharWrapper) {
	m := make([]float64, attributes.EndStatType)
	m[attributes.CD] = 0.4
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBaseWithHitlag("faruzan-c6", 240),
		Amount: func(atk *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
			if atk.Info.Element != attributes.Anemo {
				return nil, false
			}
			return m, true
		},
	})
}

const c6ICDKey = "faruzan-c6-icd"

func (c *char) c6Collapse() {
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		if dmg := args[2].(float64); dmg == 0 {
			return false
		}
		atk := args[1].(*combat.AttackEvent)
		char := c.Core.Player.ActiveChar()
		if char.Index != atk.Info.ActorIndex {
			return false
		}
		if !char.StatusIsActive(burstBuffKey) {
			return false
		}
		if c.StatusIsActive(c6ICDKey) {
			return false
		}
		c.AddStatus(c6ICDKey, 180, false)
		enemy := args[0].(*enemy.Enemy)
		c.pressurizedCollapse(enemy.Pos())
		return false
	}, "faruzan-c6-hook")
}
