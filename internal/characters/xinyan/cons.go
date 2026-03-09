package xinyan

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const c1ICDKey = "xinyan-c1-icd"

// 会心時、辛炎の通常攻撃と重撃の攻撃速度が5秒間12%上昇する。
// 5秒に1回のみ発動可能。
func (c *char) makeC1CB() combat.AttackCBFunc {
	if c.Base.Cons < 1 {
		return nil
	}
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if c.Core.Player.Active() != c.Index {
			return
		}
		if !a.IsCrit {
			return
		}
		if c.StatusIsActive(c1ICDKey) {
			return
		}
		c.AddStatus(c1ICDKey, 5*60, true)

		m := make([]float64, attributes.EndStatType)
		m[attributes.AtkSpd] = 0.12
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("xinyan-c1", 5*60),
			AffectedStat: attributes.AtkSpd,
			Amount: func() ([]float64, bool) {
				if c.Core.Player.CurrentState() != action.NormalAttackState && c.Core.Player.CurrentState() != action.ChargeAttackState {
					return nil, false
				}
				return m, true
			},
		})
	}
}

// 叛逆の弾き語りの物理ダメージ会心率が100%上昇し、発動時にシールドレベル3：熱狂のシールドを生成する。
func (c *char) c2() {
	c.c2Buff = make([]float64, attributes.EndStatType)
	c.c2Buff[attributes.CR] = 1

	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("xinyan-c2", -1),
		Amount: func(atk *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagElementalBurst {
				return nil, false
			}
			if atk.Info.Element != attributes.Physical {
				return nil, false
			}
			return c.c2Buff, true
		},
	})
}

// 情熱伝導の振りダメージが敵の物理耐性を15%減少させる。12秒間持続。
func (c *char) makeC4CB() combat.AttackCBFunc {
	if c.Base.Cons < 4 {
		return nil
	}
	return func(a combat.AttackCB) {
		e, ok := a.Target.(*enemy.Enemy)
		if !ok {
			return
		}
		e.AddResistMod(combat.ResistMod{
			Base:  modifier.NewBaseWithHitlag("xinyan-c4", 12*60),
			Ele:   attributes.Physical,
			Value: -0.15,
		})
	}
}

// 辛炎の重撃のスタミナ消費が30%減少する。さらに、重撃の攻撃力に防御力の50%分が加算される。
// func (c *char) c6() {}
