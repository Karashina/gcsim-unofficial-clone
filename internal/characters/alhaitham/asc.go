package alhaitham

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const a1IcdKey = "alhaitham-a1-icd"

// アルハイゼムの重撃または落下攻撃が敵に命中した場合、琢光鏡を1枚生成する。
// この効果は12秒ごとに1回のみ発動可能。
func (c *char) makeA1CB() combat.AttackCBFunc {
	if c.Base.Ascension < 1 {
		return nil
	}
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		// 投影がICD中の場合は無視
		if c.Core.Status.Duration(a1IcdKey) > 0 {
			return
		}

		c.Core.Status.Add(a1IcdKey, 720) // 12s
		c.mirrorGain(1)
	}
}

// アルハイゼムの元素熟知1ポイントごとに、投影攻撃と
// 殊境・顕象結縛のダメージが0.1%増加する。
// 両能力の最大ダメージ増加は100%。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("alhaitham-a4", -1),
		Amount: func(atk *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
			// 投影攻撃と元素爆発ダメージでのみ発動
			if atk.Info.AttackTag != attacks.AttackTagElementalBurst &&
				atk.Info.ICDGroup != attacks.ICDGroupAlhaithamProjectionAttack {
				return nil, false
			}

			m[attributes.DmgP] = 0.001 * c.Stat(attributes.EM)
			if m[attributes.DmgP] > 1 {
				m[attributes.DmgP] = 1
			}
			return m, true
		},
	})
}
