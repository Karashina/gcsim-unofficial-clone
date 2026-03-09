package lyney

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// リネがマジック弾を発射する際にHPを消費した場合、
// その矢が召喚するグリンマルキンハットが敵に命中すると、
// リネのエネルギーを3回復し、攻撃力の80%分ダメージが追加される。
func (c *char) makeA1CB(hpDrained bool) combat.AttackCBFunc {
	if c.Base.Ascension < 1 || !hpDrained {
		return nil
	}
	done := false
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if done {
			return
		}
		done = true
		c.AddEnergy("lyney-a1", 3)
	}
}

func (c *char) addA1(ai *combat.AttackInfo, hpDrained bool) {
	if c.Base.Ascension < 1 || !hpDrained {
		return
	}
	ai.Mult += 0.8
}

// 炎元素の影響を受けた敵に対してリネが与えるダメージに以下のバフが適用される:
// - 与えるダメージが60%増加。
// - リネ以外の炎元素パーティメンバー1人につき、さらに20%増加。
// この方法で炎元素の影響を受けた敵に対して最大100%のダメージ増加が可能。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}

	// チーム内の炎元素キャラクターを数える
	pyroCount := 0
	for _, char := range c.Core.Player.Chars() {
		if char.Base.Element == attributes.Pyro {
			pyroCount++
		}
	}

	// 固有天賦4のダメージ%値を計算
	a4Dmg := 0.6 + float64(pyroCount-1)*0.2
	if a4Dmg > 1 {
		a4Dmg = 1
	}

	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = a4Dmg
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("lyney-a4", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			r, ok := t.(core.Reactable)
			if !ok {
				return nil, false
			}
			if !r.AuraContains(attributes.Pyro) {
				return nil, false
			}
			return m, true
		},
	})
}
