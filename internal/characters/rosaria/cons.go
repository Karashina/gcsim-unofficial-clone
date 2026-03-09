package rosaria

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

// ロサリアが会心ヒット時、攻撃速度10%と通常攻撃ダメージ10%が4秒間アップ（シールド敵にも発動可能）
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
		if c.Core.Player.Active() != c.Index {
			return
		}

		m := make([]float64, attributes.EndStatType)
		m[attributes.DmgP] = 0.1
		c.AddAttackMod(character.AttackMod{
			Base: modifier.NewBaseWithHitlag("rosaria-c1-dmg", 240), // 4s
			Amount: func(atk *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
				if atk.Info.AttackTag != attacks.AttackTagNormal {
					return nil, false
				}
				return m, true
			},
		})

		mAtkSpd := make([]float64, attributes.EndStatType)
		mAtkSpd[attributes.AtkSpd] = 0.1
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("rosaria-c1-speed", 240), // 4s
			AffectedStat: attributes.AtkSpd,
			Amount: func() ([]float64, bool) {
				if c.Core.Player.CurrentState() != action.NormalAttackState {
					return nil, false
				}
				return mAtkSpd, true
			},
		})
	}
}

// 告解の兄姉の会心ヒット時、ロサリアのHPを5回復する。告解の兄姉発動ごとに1回のみ発動。
func (c *char) makeC4CB() combat.AttackCBFunc {
	if c.Base.Cons < 4 {
		return nil
	}
	done := false
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
		if done {
			return
		}
		done = true
		c.AddEnergy("rosaria-c4", 5)
	}
}

// 教典のレクイエムの攻撃が敵の物理耐性を20%低下させる（10秒間）。
func (c *char) makeC6CB() combat.AttackCBFunc {
	if c.Base.Cons < 6 {
		return nil
	}
	return func(a combat.AttackCB) {
		e, ok := a.Target.(*enemy.Enemy)
		if !ok {
			return
		}
		e.AddResistMod(combat.ResistMod{
			Base:  modifier.NewBaseWithHitlag("rosaria-c6", 600),
			Ele:   attributes.Physical,
			Value: -0.2,
		})
	}
}
