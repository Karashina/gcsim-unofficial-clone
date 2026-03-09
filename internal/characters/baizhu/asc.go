package baizhu

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/reactions"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 白朮はアクティブキャラクターの現在HPに応じて異なる効果を得る：
// ・HPが50%未満の場合、白朮は与える治療効果+20%を獲得する。
// ・HPが50%以上の場合、白朮は草元素ダメージ+25%を獲得する。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}

	// 治療部分
	mHeal := make([]float64, attributes.EndStatType)
	mHeal[attributes.Heal] = 0.2
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("baizhu-a1-heal-bonus", -1),
		AffectedStat: attributes.Heal,
		Amount: func() ([]float64, bool) {
			active := c.Core.Player.ActiveChar()
			if active.CurrentHPRatio() < 0.5 {
				return mHeal, true
			}
			return nil, false
		},
	})

	// 草元素ダメージ部分
	mDendroP := make([]float64, attributes.EndStatType)
	mDendroP[attributes.DendroP] = 0.25
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("baizhu-a1-dendro-dmg", -1),
		AffectedStat: attributes.DendroP,
		Amount: func() ([]float64, bool) {
			active := c.Core.Player.ActiveChar()
			if active.CurrentHPRatio() >= 0.5 {
				return mDendroP, true
			}
			return nil, false
		},
	})
}

// 継ぎ目なきシールドで回復したキャラクターは「草木の恩恵」効果を獲得する：
// 白朮のHP上限1,000ごと（50,000を超えない範囲）につき、燃焼・開花・超開花・烈開花の
// 反応ダメージが2%増加し、激化・拡散の反応ダメージが0.8%増加する。
//
//	この効果は6秒間持続する。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	c.Core.Player.ActiveChar().AddReactBonusMod(character.ReactBonusMod{
		Base: modifier.NewBaseWithHitlag("baizhu-a4", 6*60),
		Amount: func(ai combat.AttackInfo) (float64, bool) {
			limitHP := c.MaxHP() / 1000.0
			if limitHP > 50 {
				limitHP = 50
			}
			if ai.Catalyzed && (ai.CatalyzedType == reactions.Aggravate || ai.CatalyzedType == reactions.Spread) {
				return limitHP * 0.008, false
			}
			switch ai.AttackTag {
			case attacks.AttackTagBloom:
			case attacks.AttackTagHyperbloom:
			case attacks.AttackTagBurgeon:
			case attacks.AttackTagBurningDamage:
			default:
				return 0, false
			}

			return limitHP * 0.02, false
		},
	})
}
