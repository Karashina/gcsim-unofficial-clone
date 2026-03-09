package gorou

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 戦陣の誉を使用後、12秒間付近の全パーティメンバーの防御力が25%増加する。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	for _, char := range c.Core.Player.Chars() {
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(a1Key, 720),
			AffectedStat: attributes.DEFP,
			Amount: func() ([]float64, bool) {
				return c.a1Buff, true
			},
		})
	}
}

// ゴローは防御力に基づいて以下のダメージボーナスを受ける:
//
// - 犬坂鐌繰の昭: スキルダメージが防御力の156%分増加
func (c *char) a4Skill() float64 {
	if c.Base.Ascension < 4 {
		return 0
	}
	return c.TotalDef(false) * 1.56
}

// ゴローは防御力に基づいて以下のダメージボーナスを受ける:
//
// - 戦陣の誉: スキルダメージとCrystal Collapseダメージが防御力の15.6%分増加
func (c *char) a4Burst() float64 {
	if c.Base.Ascension < 4 {
		return 0
	}
	return c.TotalDef(false) * 0.156
}
