package amber

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 灼熱の雨の会心率を10%上昇させ、AoEを30%拡大する。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	// 会心
	m := make([]float64, attributes.EndStatType)
	m[attributes.CR] = .1
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("amber-a1", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			return m, atk.Info.AttackTag == attacks.AttackTagElementalBurst
		},
	})
	// AoE拡大
	c.burstRadius *= 1.3
}

// 狙い撃ちが弱点に命中すると攻撃力が15%上昇する（10秒間）。
func (c *char) makeA4CB() combat.AttackCBFunc {
	if c.Base.Ascension < 4 {
		return nil
	}
	done := false
	return func(a combat.AttackCB) {
		if !a.AttackEvent.Info.HitWeakPoint {
			return
		}
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if done {
			return
		}
		done = true

		m := make([]float64, attributes.EndStatType)
		m[attributes.ATKP] = 0.15
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("amber-a4", 600),
			AffectedStat: attributes.ATKP,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}
}
