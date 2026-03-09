package albedo

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	hexereiBuffKey      = "albedo-hexerei-dmg"
	hexereiBuffDuration = 20 * 60 // 20s
)

// 創生術・擬似陽花(Solar Isotoma)が生成するTransient Blossomは、
// HPが50%以下の敵に対して与えるダメージが25%増加する。
func (c *char) a1() {
	if !c.Core.Combat.DamageMode {
		return
	}
	if c.Base.Ascension < 1 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = 0.25
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("albedo-a1", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagElementalArt {
				return nil, false
			}
			// 更新時に自身ではトリガーされない
			if atk.Info.Abil == "Abiogenesis: Solar Isotoma" {
				return nil, false
			}
			if e, ok := t.(*enemy.Enemy); !(ok && e.HP()/e.MaxHP() < .5) {
				return nil, false
			}
			return m, true
		},
	})
}

// 誕生式・大地の潮を使用すると、周囲のパーティメンバーの元素熟知が10秒間125増加する。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.EM] = 125
	for _, char := range c.Core.Player.Chars() {
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("albedo-a4", 600),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}
}

// Hexerei: Secret Rite
// Solar Isotoma生成後、周囲のパーティメンバーのダメージが防御力1000あたり4%増加（最大12%）。
// Silver Isotoma生成後、Hexereiパーティメンバーのダメージが防御力1000あたり10%増加（最大30%）。
func (c *char) hexereiSecretRite() {
	// 防御力ベースのボーナスを計算
	def := c.TotalDef(false)
	defUnits := def / 1000.0

	// Hexereiステータスに基づいて異なるバフを適用
	for _, char := range c.Core.Player.Chars() {
		// Conditionを照会してこのキャラクターがHexereiかチェック
		isHexerei := false
		if result, err := char.Condition([]string{"hexerei"}); err == nil {
			if val, ok := result.(bool); ok && val {
				isHexerei = true
			}
		}

		var dmgBonus float64
		var maxBonus float64

		if c.isHexerei && isHexerei {
			// Silver Isotoma (Hexerei Albedo) -> Hexereiパーティメンバーに強化バフ
			dmgBonus = 0.10 * defUnits // 防御力1000あたり10%
			maxBonus = 0.30            // 最大30%
		} else {
			// Solar Isotoma -> 全パーティメンバーに通常バフ
			dmgBonus = 0.04 * defUnits // 防御力1000あたり4%
			maxBonus = 0.12            // 最大12%
		}

		// 最大値で制限
		if dmgBonus > maxBonus {
			dmgBonus = maxBonus
		}

		finalBonus := dmgBonus
		m := make([]float64, attributes.EndStatType)
		m[attributes.DmgP] = finalBonus

		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBaseWithHitlag(fmt.Sprintf("%s-%d", hexereiBuffKey, char.Index), hexereiBuffDuration),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				switch atk.Info.AttackTag {
				case attacks.AttackTagNormal,
					attacks.AttackTagExtra,
					attacks.AttackTagPlunge,
					attacks.AttackTagElementalArt,
					attacks.AttackTagElementalArtHold,
					attacks.AttackTagElementalBurst:
					return m, true
				}
				return nil, false
			},
		})
	}

	c.Core.Log.NewEvent("Albedo Hexerei: Secret Rite applied", glog.LogCharacterEvent, c.Index).
		Write("def", def).
		Write("is_hexerei", c.isHexerei)
}
