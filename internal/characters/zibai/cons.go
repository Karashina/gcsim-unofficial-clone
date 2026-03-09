package zibai

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// c1Init は1凸を初期化する: Burst Forth With Vigor, But Enter in Silence
// 元素スキル使用後、ジバイは即座に月相転移輝度を100獲得し、
// 月相転移モードごとの神馬駆けの最大使用回数が5回に増加する。
// さらに、月相転移モードに切り替わるたびに、初回の神馬駆けの
// 2段目のLunar-Crystallize反応ダメージが220%増加する。
func (c *char) c1Init() {
	c.maxSpiritSteedUsages = 5

	c.Core.Log.NewEvent("Zibai C1 active: Max Spirit Steed usages increased to 5", glog.LogCharacterEvent, c.Index)
}

// c2Init は2凸を初期化する: At Birth Are Souls Born, and in Death Leave But Husks
// 月相転移モード中、近くの全パーティメンバーのLunar-Crystallize反応ダメージが30%増加する。
// 月相がAscendant Gleamの時、固有天賦1が強化され、
// 神馬駆けの2段目のダメージがジバイの防御力の550%分さらに増加する。
// 固有天賦1のアンロックが必要。
func (c *char) c2Init() {
	// 月相転移中、全パーティメンバーにLCrs反応ボーナスを追加
	for _, char := range c.Core.Player.Chars() {
		char.AddLCrsReactBonusMod(character.LCrsReactBonusMod{
			Base: modifier.NewBase("zibai-c2-lcrs-bonus", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				if !c.lunarPhaseShiftActive {
					return 0, false
				}
				return 0.30, false
			},
		})
	}

	c.Core.Log.NewEvent("Zibai C2 active: Party LCrs bonus +30% during Lunar Phase Shift", glog.LogCharacterEvent, c.Index)
}

// c4Init は4凸を初期化する: The Spirit Passes, Then Form Follows
// 月相転移モード中、ジバイの通常攻撃シーケンスがリセットされなくなり、
// 神馬駆けが敵に命中すると、ジバイは「Scattermoon Splendor」効果を獲得する:
// 次に通常攻撃を使用する際、4段目の追加攻撃が元のダメージの
// 250%のLunar-Crystallize反応ダメージを与える。
func (c *char) c4Init() {
	// Scattermoon SplendorはspiritSteedOnHitCBとqueueN4AdditionalHitで処理
	// 通常攻撃シーケンスのリセット無効化はattack.goのsavedNormalCounterで処理

	c.Core.Log.NewEvent("Zibai C4 active: Scattermoon Splendor and Normal Attack persistence enabled", glog.LogCharacterEvent, c.Index)
}

// c6Init は6凸を初期化する: The World, A Journey in Passing
// ジバイが月相転移モード中、月相転移輝度の獲得率が50%増加する。
// さらに、神馬駆けは全ての月相転移輝度を消費するように変更される。
// これにより、この神馬駆けのダメージと次の3秒間のジバイのLunar-Crystallize
// 反応ダメージが70を超えた1ポイントごとに1.6%上昇する。
// この効果は重複しない。
func (c *char) c6Init() {
	// 50%輝度獲得増加はaddPhaseShiftRadianceで処理
	// 全輝度消費とelevationバフはspiritSteedStrideで処理

	c.Core.Log.NewEvent("Zibai C6 active: Enhanced radiance gain and elevation buff", glog.LogCharacterEvent, c.Index)
}

// applyC6ElevationBuff は6凸のelevダメージバフを適用する
func (c *char) applyC6ElevationBuff(bonusPct float64) {
	const c6Duration = 3 * 60 // 3 seconds

	c.AddStatus(c6ElevationBuffKey, c6Duration, true)

	// 神馬駆けとLCrsダメージ用のelev modを追加
	c.AddElevationMod(character.ElevationMod{
		Base: modifier.NewBaseWithHitlag(c6ElevationBuffKey, c6Duration),
		Amount: func(ai combat.AttackInfo) (float64, bool) {
			// ジバイからの神馬駆けとLCrs攻撃にのみ適用
			if ai.ActorIndex != c.Index {
				return 0, false
			}
			if ai.AttackTag != attacks.AttackTagElementalArt &&
				ai.AttackTag != attacks.AttackTagLCrsDamage {
				return 0, false
			}
			return bonusPct, false
		},
	})

	c.Core.Log.NewEvent("Zibai C6 Elevation buff applied", glog.LogCharacterEvent, c.Index).
		Write("bonus_pct", bonusPct).
		Write("duration", c6Duration)
}
