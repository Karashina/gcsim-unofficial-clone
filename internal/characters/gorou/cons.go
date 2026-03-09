package gorou

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/shield"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 1命ノ星座:
// ゴローの「大将旗」またはGeneral's GloryのAoE内で、
// ゴロー以外のキャラクターが敵に岩元素ダメージを与えると、
// ゴローの「犬坂鐌繰の昭」のCDが2秒減少する。この効果は10秒に1回のみ発動。
func (c *char) c1() {
	icd := -1
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		eActive := c.Core.Status.Duration(generalWarBannerKey) > 0
		qActive := c.Core.Status.Duration(generalGloryKey) > 0
		if !eActive && !qActive {
			return false
		}
		if icd > c.Core.F {
			return false
		}

		trg := args[0].(combat.Target)
		// ターゲットがフィールド内か確認が必要
		var area combat.AttackPattern
		if eActive {
			area = c.eFieldArea
		} else {
			// 元素スキルと元素爆発は同時に存在できない
			// 元素爆発がアクティブなら、範囲は現在のプレイヤー位置周辺
			area = combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 8)
		}
		if !trg.IsWithinArea(area) {
			return false
		}

		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex == c.Index {
			return false
		}
		if atk.Info.Element != attributes.Geo {
			return false
		}

		dmg := args[2].(float64)
		if dmg == 0 {
			return false
		}

		icd = c.Core.F + 600
		c.ReduceActionCooldown(action.ActionSkill, 120)
		return false
	}, "gorou-c1")
}

// 2命ノ星座:
// General's Gloryが効果中、付近のアクティブキャラが結晶反応で
// 元素結晶片を取得すると、持続時間が1秒延長される。
// この効果は0.1秒に1回のみ発動。最大延長は3秒。
func (c *char) c2() {
	c.Core.Events.Subscribe(event.OnShielded, func(args ...interface{}) bool {
		if c.Core.Status.Duration(generalGloryKey) <= 0 {
			return false
		}
		// シールドをチェック
		shd := args[0].(shield.Shield)
		if shd.Type() != shield.Crystallize {
			return false
		}
		if c.c2Extension >= 3 {
			return false
		}
		c.c2Extension++
		c.Core.Status.Extend(generalGloryKey, 60)
		return false
	}, "gorou-c2")
}

// 6命ノ星座:
// 犬坂鐌繰の昭または戦陣の誉を使用してから12秒間、
// 使用時のスキルフィールドのバフレベルに応じて、
// 付近の全パーティメンバーの岩元素ダメージの会心ダメージが増加:
// • 「積石」: +10%
// • 「集岩」: +20%
// • 「碎岩」: +40%
// この効果は重複せず、最後に発動したインスタンスを参照する。
func (c *char) c6() {
	for _, char := range c.Core.Player.Chars() {
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBaseWithHitlag(c6key, 720),
			Amount: func(ae *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
				if ae.Info.Element != attributes.Geo {
					return nil, false
				}
				return c.c6Buff, true
			},
		})
	}
}
