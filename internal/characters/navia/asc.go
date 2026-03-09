package navia

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	a1Key = "navia-a1-dmg"
)

// セレモニアルクリスタルショット使用後4秒間、ナヴィアの通常攻撃、
// 重撃、落下攻撃ダメージが岩元素ダメージに変換され（他の元素付与で上書き不可）、
// ナヴィアの通常攻撃、重撃、落下攻撃ダメージが40%増加する。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	c.Core.Log.NewEvent("a1 infusion added", glog.LogCharacterEvent, c.Index)

	// ダメージボーナスを追加
	m := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBaseWithHitlag(a1Key, 60*4), // 4s
		Amount: func(atk *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
			// 通常攻撃/重撃/落下攻撃以外はスキップ
			if atk.Info.AttackTag != attacks.AttackTagNormal &&
				atk.Info.AttackTag != attacks.AttackTagExtra &&
				atk.Info.AttackTag != attacks.AttackTagPlunge {
				return nil, false
			}
			// バフを適用
			m[attributes.DmgP] = 0.4
			return m, true
		},
	})
}

// 烤/雷/氷/水元素のパーティメンバーごとに、ナヴィアの攻撃力が20%増加。
// この効果は最大2回まで重複可能。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}

	ele := 0
	for _, char := range c.Core.Player.Chars() {
		switch char.Base.Element {
		case attributes.Pyro, attributes.Electro, attributes.Cryo, attributes.Hydro:
			ele++
		}
	}
	if ele > 2 {
		ele = 2
	}

	m := make([]float64, attributes.EndStatType)
	m[attributes.ATKP] = 0.2 * float64(ele)
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("navia-a4", -1),
		AffectedStat: attributes.ATKP,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})
}
