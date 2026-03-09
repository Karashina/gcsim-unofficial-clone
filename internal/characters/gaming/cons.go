package gaming

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const c2Key = "gaming-c2"
const c4Key = "gaming-c4"
const c6Key = "gaming-c6"

// 獣神金舞の獣神マンチャイが嘉明のもとに戻ったとき、
// 嘉明のHPの15%を回復する。
func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}

	c.Core.Player.Heal(info.HealInfo{
		Caller:  c.Index,
		Target:  c.Index,
		Message: "Bringer of Blessing (C1)",
		Type:    info.HealTypePercent,
		Src:     0.15,
		Bonus:   c.Stat(attributes.Heal),
	})
}

// 嘉明が治療を受けた際、その治療が超過回復した場合、
// 攻撃力+20%、5秒間。
func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}

	m := make([]float64, attributes.EndStatType)
	m[attributes.ATKP] = 0.2
	c.Core.Events.Subscribe(event.OnHeal, func(args ...interface{}) bool {
		hi := args[0].(*info.HealInfo)
		overheal := args[3].(float64)

		if overheal <= 0 {
			return false
		}

		if hi.Target != c.Index {
			return false
		}

		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(c2Key, 5*60),
			AffectedStat: attributes.ATKP,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})

		return false
	}, c2Key+"-on-heal")
}

// 百獣郡舞の落下攻撃・祥雲の幸矢が敵に命中すると、
// 嘉明の元素エネルギーを2回復。この効果は0.2秒ごとに1回発動可能。
func (c *char) makeC4CB() combat.AttackCBFunc {
	if c.Base.Cons < 4 {
		return nil
	}
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if c.StatusIsActive(c4Key) {
			return
		}
		c.AddStatus(c4Key, 0.2*60, true)
		c.AddEnergy(c4Key, 2)
	}
}

// 百獣郡舞の落下攻撃・祥雲の幸矢の会心率+20%、会心ダメージ+40%、
// 攻撃範囲が拡大される。
func (c *char) c6() {
	if c.Base.Cons < 6 {
		c.specialPlungeRadius = 4
		return
	}
	c.specialPlungeRadius = 6

	m := make([]float64, attributes.EndStatType)
	m[attributes.CR] = 0.2
	m[attributes.CD] = 0.4
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase(c6Key, -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.Abil != specialPlungeKey {
				return nil, false
			}
			return m, true
		},
	})
}
