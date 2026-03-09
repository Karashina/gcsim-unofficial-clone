package nefer

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 4凸：オンフィールド時に翠露獲得+25%、Shadow Dance中は周囲の敵の草元素耐性-20%（離脱後4.5秒で解除）
func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}
	// オンフィールド制約を追加
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(args ...interface{}) bool {
		prev := args[0].(int)
		next := args[1].(int)

		if prev == c.Index && c.c4buffkey {
			c.Core.Player.Verdant.SetGainBonus(c.Core.Player.Verdant.GetGainBonus() - 0.25)
			c.c4buffkey = false
		}
		if next == c.Index {
			c.Core.Player.Verdant.SetGainBonus(c.Core.Player.Verdant.GetGainBonus() + 0.25)
			c.c4buffkey = true
		}
		return false
	}, "nefer-c4-verdantdewgain")

	// Shadow Dance中に周囲の敵に草元素耐性-20%を適用。
	// NeferがShadow Danceを離脱した場合、この効果は追加で4.5秒間持続する。
	// Neferがアクティブな時にスキル使用を検出し、周囲の敵に耐性モディファイアを適用する。
	c.Core.Events.Subscribe(event.OnSkill, func(args ...interface{}) bool {
		// Neferがアクティブキャラクターの場合のみ発動
		if c.Core.Player.Active() != c.Index {
			return false
		}

		// 半径：「周囲」として10ユニットを使用（他のスキルでの一般的な規定）
		enemies := c.Core.Combat.EnemiesWithinArea(combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 10), nil)
		if len(enemies) == 0 {
			return false
		}

		// 持続時間：Shadow Dance基本10秒(10*60) + 4.5秒(270フレーム) = 870フレーム
		dur := 10*60 + 270
		for _, e := range enemies {
			targ, ok := e.(*enemy.Enemy)
			if !ok {
				continue
			}
			targ.AddResistMod(combat.ResistMod{
				Base:  modifier.NewBaseWithHitlag("nefer-c4-dendro", dur),
				Ele:   attributes.Dendro,
				Value: -0.20,
			})
		}

		return false
	}, "nefer-c4")
}

// 6凸：PPの第2ヒットを変換し、追加の範囲LBダメージ（EMスケーリング）を追加。ムーンサインAscendantならLBダメージ+15%（ElevationMod経由）。
func (c *char) c6() {
	if c.Base.Cons < 6 {
		return
	}

	// ムーンサイン：Ascendant Gleam - Lunar-BloomダメージのElevationを15%増加
	// キャラクターシステムのElevationBonusで処理
	if c.MoonsignAscendant {
		c.AddElevationMod(character.ElevationMod{
			Base: modifier.NewBase("Nefer C6", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				if ai.AttackTag == attacks.AttackTagLBDamage {
					return 0.15, false
				}
				return 0, false
			},
		})
	}
}
