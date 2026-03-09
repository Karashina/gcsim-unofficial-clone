package kazuha

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 2凸:
// 千早振るが生成する秋の旋風フィールドは以下の効果を持つ:
// - 万葉の狂風と千早振る自身の元素熟知+200。
// - フィールド内のキャラクターの元素熟知+200。
// この元素熟知増加効果は重複しない。
func (c *char) c2(src int) func() {
	return func() {
		// srcが変更されていたらティックしない
		if c.qFieldSrc != src {
			c.Core.Log.NewEvent("kazuha q src check ignored, src diff", glog.LogCharacterEvent, c.Index).
				Write("src", src).
				Write("new src", c.qFieldSrc)
			return
		}
		// 元素爆発が終了していたらティックしない
		if c.Core.Status.Duration(burstStatus) == 0 {
			return
		}

		// 0.5秒後に再チェック
		c.Core.Tasks.Add(c.c2(src), 30)

		ap := combat.NewCircleHitOnTarget(c.qAbsorbCheckLocation.Shape.Pos(), nil, 9)
		if !c.Core.Combat.Player().IsWithinArea(ap) {
			return
		}

		// 元素爆発範囲内ならバフを適用
		c.Core.Log.NewEvent("kazuha-c2 ticking", glog.LogCharacterEvent, -1)

		// アクティブキャラに2凸バフを1秒間適用
		active := c.Core.Player.ActiveChar()
		active.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("kazuha-c2", 60), // 1s
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				return c.c2buff, true
			},
		})

		// 万葉の狂風にも（フィールド外でも）バフを1秒間適用
		if active.Base.Key != c.Base.Key {
			c.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag("kazuha-c2", 60), // 1s
				AffectedStat: attributes.EM,
				Amount: func() ([]float64, bool) {
					return c.c2buff, true
				},
			})
		}
	}
}

// 6凸
// 万葉の狂風または千早振るを使用後、風元素付与を5秒間得る。
// さらに元素熟知1あたり、通常攻撃・重撃・落下攻撃の
// ダメージが0.2%増加する。
func (c *char) c6() {
	// 風元素付与を追加
	c.Core.Player.AddWeaponInfuse(
		c.Index,
		"kazuha-c6-infusion",
		attributes.Anemo,
		60*5,
		true,
		attacks.AttackTagNormal, attacks.AttackTagExtra, attacks.AttackTagPlunge,
	)
	c.Core.Events.Emit(event.OnInfusion, c.Index, attributes.Anemo, 60*5)

	// 元素熟知ベースのバフを追加
	m := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBaseWithHitlag("kazuha-c6-dmgup", 60*5), // 5s
		Amount: func(atk *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
			// 通常攻撃/重撃/落下攻撃以外はスキップ
			if atk.Info.AttackTag != attacks.AttackTagNormal &&
				atk.Info.AttackTag != attacks.AttackTagExtra &&
				atk.Info.AttackTag != attacks.AttackTagPlunge {
				return nil, false
			}
			// バフを適用
			m[attributes.DmgP] = 0.002 * c.Stat(attributes.EM)
			return m, true
		},
	})
}
