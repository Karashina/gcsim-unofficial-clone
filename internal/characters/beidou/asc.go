package beidou

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 固有天賦1は未実装:
// TODO: 踏潮の攻撃が攻撃を受けた瞬間にカウンターを行うと、最大ダメージボーナスを獲得する。

// 踏潮の最大ダメージボーナスで発動後10秒間、以下の効果を獲得:
// - 通常攻撃と重撃のダメージが15%増加。通常攻撃と重撃の攻撃速度が15%増加。
// TODO: - 重撃の発動前の遅延が大幅に短縮。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}

	mDmg := make([]float64, attributes.EndStatType)
	mDmg[attributes.DmgP] = .15
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBaseWithHitlag("beidou-a4-dmg", 600),
		Amount: func(atk *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagNormal && atk.Info.AttackTag != attacks.AttackTagExtra {
				return nil, false
			}
			return mDmg, true
		},
	})

	mAtkSpd := make([]float64, attributes.EndStatType)
	mAtkSpd[attributes.AtkSpd] = .15
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag("beidou-a4-atkspd", 600),
		AffectedStat: attributes.AtkSpd,
		Amount: func() ([]float64, bool) {
			return mAtkSpd, true
		},
	})
}
