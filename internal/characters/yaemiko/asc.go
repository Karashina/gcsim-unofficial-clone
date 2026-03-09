package yaemiko

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 大秘法・天狐顯真を発動時、破壊された殺生櫻1本につき
// 野干玉の創出・殺生櫻のCDを1回分リセットする。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	c.ResetActionCooldown(action.ActionSkill)
}

// 八重神子が持つ元素熟知1ポイントごとに殺生櫻のダメージが0.15%増加する。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("yaemiko-a4", -1),
		Amount: func(atk *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
			// 元素スキルダメージのみトリガー
			if atk.Info.AttackTag != attacks.AttackTagElementalArt {
				return nil, false
			}
			m[attributes.DmgP] = c.Stat(attributes.EM) * 0.0015
			return m, true
		},
	})
}
