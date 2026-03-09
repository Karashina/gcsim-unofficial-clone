package columbina

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/shield"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	// 1凸
	c1Key            = "c1-gravity-interference"
	c1GravitySkipKey = "c1-gravity-skip"
	c1ICD            = 15 * 60 // 15s ICD
	c1Elevation      = 0.015   // 1.5% Elevation bonus for all party members

	// 2凸
	c2Key = "lunar-brilliance"
	c2Dur = 8 * 60 // 8s duration

	// 4凸
	c4IcdKey      = "c4-gravity-bonus-icd"
	c4Energy      = 4
	c4HPBonusLC   = 0.125 // 最大HPの12.5%
	c4HPBonusLB   = 0.025 // 最大HPの2.5%
	c4HPBonusLCrs = 0.125 // 最大HPの12.5%

	// 6凸
	c6Key       = "columbina-c6-crit-dmg"
	c6Dur       = 8 * 60 // 8s duration
	c6CDBonus   = 0.80   // 80% CRIT DMG
	c6Elevation = 0.07   // 7% Elevation
)

// 1凸: スキル発動時にGravity Interference効果をトリガー（15sに1回）
// 効果は支配的タイプに基づき、エネルギー回復、耿力回復、またはシールドを提供
// また全パーティメンバーにLunar反応ダメージ1.5% Elevationボーナスを提供
func (c *char) c1Init() {
	if c.Base.Cons < 1 {
		return
	}

	// 全パーティメンバーにLunar反応ダメージのみに1.5% Elevationボーナスを適用
	// precalc（calcLunarChargedDmg等）で評価され、atkには正しいAttackTagが設定されている
	for _, char := range c.Core.Player.Chars() {
		char.AddElevationMod(character.ElevationMod{
			Base: modifier.NewBase("columbina-c1-elevation", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				if ai.AttackTag == attacks.AttackTagLCDamage ||
					ai.AttackTag == attacks.AttackTagLBDamage ||
					ai.AttackTag == attacks.AttackTagLCrsDamage {
					return c1Elevation, false
				}
				return 0, false
			},
		})
	}
}

func (c *char) c1OnSkill() {
	if c.Base.Cons < 1 {
		return
	}
	if c.StatusIsActive(c1Key) {
		return
	}

	c.AddStatus(c1Key, c1ICD, false)

	dominantType := c.getDominantLunarType()

	c.Core.Log.NewEvent("C1 Gravity Interference triggered on skill", glog.LogCharacterEvent, c.Index).
		Write("dominant_type", dominantType)

	c.AddStatus(c1GravitySkipKey, -1, false)
	c.triggerGravityInterference()

}

func (c *char) c1OnGravityInterference() {
	if c.Base.Cons < 1 {
		return
	}

	dominantType := c.getDominantLunarType()

	switch dominantType {
	case "lc":
		// エネルギー回復
		c.AddEnergy("c1-energy", 6)
	case "lcrs":
		// Rainsea Shield召喚: HP上限の12%、水元素ダメージに250%有効性、8s持続
		shieldAmount := c.MaxHP() * 0.12
		// Rainsea Shieldの実装
		// Rainsea Shieldを適用（水元素250%有効性、8s持続）
		importShield := func() {
			// gcsim標準のシールドAPIを使用
			// ShieldType: カスタム用にEndType+1を使用（必要なら shield.go にColumbinaShieldを定義）
			// Target: アクティブキャラクター
			c.Core.Player.Shields.Add(&shield.Tmpl{
				ActorIndex: c.Index,
				Target:     c.Core.Player.Active(),
				Name:       "Rainsea Shield",
				Src:        c.Index,
				ShieldType: shield.ColumbinaC1,
				Ele:        attributes.Hydro,
				HP:         shieldAmount,
				Expires:    c.Core.F + 8*60,
			})
			c.Core.Log.NewEvent("C1 Rainsea Shield applied", glog.LogCharacterEvent, c.Index).
				Write("amount", shieldAmount)
		}
		importShield()
	}
}

// 2凸: Gravity蓄積率が34%増加
// Gravity Interference時、Lunar Brillianceを獲得（HP上限40%、8s間）、支配的Lunarタイプに基づくステータスバフ
// 月印がAscendant Gleamの場合、支配的タイプに基づいてアクティブキャラにバフを適用:
// - LC: 攻撃力 = ColumbinaのHP上限の1%
// - LB: 元素熟知 = ColumbinaのHP上限の0.35%
// - LCrs: 防御力 = ColumbinaのHP上限の1%
func (c *char) c2OnGravityInterference(dominantType string) {
	if c.Base.Cons < 2 {
		return
	}

	// Lunar Brillianceを獲得（ColumbinaにHP上限40%ブースト8s間）
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag(c2Key+"-hp", c2Dur),
		AffectedStat: attributes.HPP,
		Amount: func() ([]float64, bool) {
			m := make([]float64, attributes.EndStatType)
			m[attributes.HPP] = 0.40 // HP上限40%
			return m, true
		},
	})

	// 月印がAscendant Gleamの場合、アクティブキャラに追加バフを適用
	if !c.MoonsignAscendant {
		return
	}

	columbinaMHP := c.Stat(attributes.HP)

	// バフは現在フィールド上のキャラのみに影響するように、
	// 全パーティメンバーにStatModを追加し、そのメンバーが
	// アクティブキャラの時のみバフ量を返す。交代時には適用されなくなる。
	switch dominantType {
	case "lc":
		// 攻撃力増加 = ColumbinaのHP上限の1%
		buffValue := 0.01 * columbinaMHP
		for _, ch := range c.Core.Player.Chars() {
			key := fmt.Sprintf("%s-atk-%d", c2Key, ch.Index)
			ch.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag(key, c2Dur),
				AffectedStat: attributes.ATK,
				Amount: func() ([]float64, bool) {
					m := make([]float64, attributes.EndStatType)
					if c.Core.Player.Active() == ch.Index {
						m[attributes.ATK] = buffValue
					}
					return m, true
				},
			})
		}
	case "lb":
		// 元素熟知増加 = ColumbinaのHP上限の0.35%
		buffValue := 0.0035 * columbinaMHP
		for _, ch := range c.Core.Player.Chars() {
			key := fmt.Sprintf("%s-em-%d", c2Key, ch.Index)
			ch.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag(key, c2Dur),
				AffectedStat: attributes.EM,
				Amount: func() ([]float64, bool) {
					m := make([]float64, attributes.EndStatType)
					if c.Core.Player.Active() == ch.Index {
						m[attributes.EM] = buffValue
					}
					return m, true
				},
			})
		}
	case "lcrs":
		// 防御力増加 = ColumbinaのHP上限の1%
		buffValue := 0.01 * columbinaMHP
		for _, ch := range c.Core.Player.Chars() {
			key := fmt.Sprintf("%s-def-%d", c2Key, ch.Index)
			ch.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag(key, c2Dur),
				AffectedStat: attributes.DEF,
				Amount: func() ([]float64, bool) {
					m := make([]float64, attributes.EndStatType)
					if c.Core.Player.Active() == ch.Index {
						m[attributes.DEF] = buffValue
					}
					return m, true
				},
			})
		}
	}

	c.Core.Log.NewEvent("C2 Lunar Brilliance activated", glog.LogCharacterEvent, c.Index).
		Write("dominant_type", dominantType).
		Write("columbina_mhp", columbinaMHP).
		Write("duration", c2Dur)
}

// 4凸: Gravity Interference時にエネルギー4を回復しHP係数ダメージボーナスを追加
// Lunar反応ダメージがLC/LB/LCrsでそれぞれHP上限の12.5%/2.5%/12.5%増加
func (c *char) c4OnGravityInterference(dominantType string) {
	if c.Base.Cons < 4 {
		return
	}

	// エネルギーを回復
	c.AddEnergy("c4-energy", c4Energy)

	// 4凸ボーナス適用用に支配的タイプを記録
	c.c4DominantType = dominantType

	c.Core.Log.NewEvent("C4 energy restored", glog.LogCharacterEvent, c.Index).
		Write("energy", c4Energy).
		Write("dominant_type", dominantType)
}

// 6凸: Lunar Domain内のキャラがLunar反応をトリガーした後8s間、
// 関与した元素に応じて対応する元素ダメージの会心ダメージが80%増加。
// 同じ元素の効果は重複しない。
// また全パーティメンバーにLunar反応ダメージ7% Elevationボーナスを提供。
func (c *char) c6Init() {
	if c.Base.Cons < 6 {
		return
	}

	// 全パーティメンバーにLunar反応ダメージ7% Elevationボーナスを適用
	for _, char := range c.Core.Player.Chars() {
		char.AddElevationMod(character.ElevationMod{
			Base: modifier.NewBase("columbina-c6-elevation", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				if ai.AttackTag == attacks.AttackTagLCDamage ||
					ai.AttackTag == attacks.AttackTagLBDamage ||
					ai.AttackTag == attacks.AttackTagLCrsDamage {
					return c6Elevation, false
				}
				return 0, false
			},
		})
	}

	// 全てのLunar反応イベントを購読して会心ダメージバフを適用
	c.Core.Events.Subscribe(event.OnLunarCharged, func(args ...interface{}) bool {
		c.c6ApplyBuffToAllChars(attributes.Electro)
		c.c6ApplyBuffToAllChars(attributes.Hydro)
		return false
	}, "columbina-c6-lc-trigger")

	c.Core.Events.Subscribe(event.OnLunarBloom, func(args ...interface{}) bool {
		c.c6ApplyBuffToAllChars(attributes.Dendro)
		c.c6ApplyBuffToAllChars(attributes.Hydro)
		return false
	}, "columbina-c6-lb-trigger")

	c.Core.Events.Subscribe(event.OnLunarCrystallize, func(args ...interface{}) bool {
		c.c6ApplyBuffToAllChars(attributes.Geo)
		c.c6ApplyBuffToAllChars(attributes.Hydro)
		return false
	}, "columbina-c6-lcrs-trigger")
}

// c6ApplyBuffToAllCharsは全パーティメンバーに対応する元素の6凸会心ダメージバフを適用
func (c *char) c6ApplyBuffToAllChars(element attributes.Element) {
	if c.Base.Cons < 6 {
		return
	}

	// 全パーティメンバーに会心ダメージバフを適用
	for _, char := range c.Core.Player.Chars() {
		if char.Base.Element.String() == element.String() {
			char.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag(c6Key, c6Dur),
				AffectedStat: attributes.CD,
				Amount: func() ([]float64, bool) {
					m := make([]float64, attributes.EndStatType)
					m[attributes.CD] = c6CDBonus
					return m, true
				},
			})
		}
	}

	c.Core.Log.NewEvent("C6 CRIT DMG buff applied to all party members", glog.LogCharacterEvent, c.Index).
		Write("element", element.String()).
		Write("crit_dmg_bonus", c6CDBonus).
		Write("duration", c6Dur)
}

// 3凸: 全周辺パーティメンバーのLunar反応ダメージが1.5% Elevation
// 5凸: 全周辺パーティメンバーのLunar反応ダメージが1.5% Elevation
// 1凸の1.5%と重複し、5凸で合計4.5%
func (c *char) c3c5Init() {
	// 3凸: +1.5% Elevation
	if c.Base.Cons >= 3 {
		for _, char := range c.Core.Player.Chars() {
			char.AddElevationMod(character.ElevationMod{
				Base: modifier.NewBase("columbina-c3-elevation", -1),
				Amount: func(ai combat.AttackInfo) (float64, bool) {
					if ai.AttackTag == attacks.AttackTagLCDamage ||
						ai.AttackTag == attacks.AttackTagLBDamage ||
						ai.AttackTag == attacks.AttackTagLCrsDamage {
						return 0.015, false
					}
					return 0, false
				},
			})
		}
	}

	// 5凸: +1.5% Elevation（3凸と重複）
	if c.Base.Cons >= 5 {
		for _, char := range c.Core.Player.Chars() {
			char.AddElevationMod(character.ElevationMod{
				Base: modifier.NewBase("columbina-c5-elevation", -1),
				Amount: func(ai combat.AttackInfo) (float64, bool) {
					if ai.AttackTag == attacks.AttackTagLCDamage ||
						ai.AttackTag == attacks.AttackTagLBDamage ||
						ai.AttackTag == attacks.AttackTagLCrsDamage {
						return 0.015, false
					}
					return 0, false
				},
			})
		}
	}
}
