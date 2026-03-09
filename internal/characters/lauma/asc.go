package lauma

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// A0
// Laumaの元素熟知1ポイントにつき、Lunar-Bloomの基礎ダメージが0.0175%増加する（最大14%）。
func (c *char) a0() {
	for _, char := range c.Core.Player.Chars() {
		char.AddStatus("LB-Key", -1, false)
		char.AddLBBaseReactBonusMod(character.LBBaseReactBonusMod{
			Base: modifier.NewBase("Moonsign Benediction: Nature's Chorus (A0)", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				maxval := 0.14
				return min(maxval, c.Stat(attributes.EM)*0.000175), false
			},
		})
	}
}

// 固有天賦1
// Laumaが元素スキルを使用した後20秒間、
// パーティのムーンサインに応じて異なるバフ効果が付与される。
// 異なるムーンサインレベルのバフは重複しない。
// ムーンサイン：Nascent Gleam
// 周囲のパーティメンバー全員の開花・超開花・烈開花ダメージが会心可能になる。
// 会心率は15%固定、会心ダメージは100%固定。
// この効果の会心率は、元素反応を会心可能にする同様の効果の会心率と加算される。
// ムーンサイン：Ascendant Gleam
// 周囲のパーティメンバー全員のLunar-BloomダメージのCR+10%、CD+20%。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}

	// 適切なムーンサインバフを20秒間適用
	if c.MoonsignNascent {
		// ムーンサイン：Nascent Gleam - 開花・超開花・烈開花ダメージが会心可能
		// 会心率15%固定、会心ダメージ100%固定
		c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
			ae := args[1].(*combat.AttackEvent)

			switch ae.Info.AttackTag {
			case attacks.AttackTagBurningDamage:
			case attacks.AttackTagBloom:
			case attacks.AttackTagHyperbloom:
			case attacks.AttackTagBurgeon:
			default:
				return false
			}

			// これらの反応に特殊会心処理を追加
			ae.Snapshot.Stats[attributes.CR] += 0.15
			ae.Snapshot.Stats[attributes.CD] += 1.0

			c.Core.Log.NewEvent("lauma a1 nascent crit buff", glog.LogCharacterEvent, ae.Info.ActorIndex).
				Write("final_crit", ae.Snapshot.Stats[attributes.CR])

			return false
		}, "lauma-a1-nascent-reaction-crit")
	} else if c.MoonsignAscendant {
		c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
			ae := args[1].(*combat.AttackEvent)

			switch ae.Info.AttackTag {
			case attacks.AttackTagLBDamage:
			default:
				return false
			}
			ae.Snapshot.Stats[attributes.CR] += 0.1
			ae.Snapshot.Stats[attributes.CD] += 0.2

			c.Core.Log.NewEvent("lauma a1 nascent crit buff", glog.LogCharacterEvent, ae.Info.ActorIndex)

			return false
		}, "lauma-a1-nascent-reaction-crit")
	}
}

// 固有天賦2
// Laumaの元素熟知1ポイントにつき以下のボーナスを得る：
// 元素スキルのダメージが0.04%増加する。この方法で得られる最大増加量は32%。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}

	// 元素スキルのダメージボーナスを追加
	m := make([]float64, attributes.EndStatType)
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("lauma-a4-skill-dmg-bonus", -1), // Permanent
		Amount: func(atk *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
			// 元素スキル以外はスキップ
			if atk.Info.AttackTag != attacks.AttackTagElementalArt {
				return nil, false
			}
			// 元素熟知ボーナスを計算
			em := c.Stat(attributes.EM)
			bonus := min(0.32, em*0.0004) // 元素熟知あたり0.04%、最大32%
			m[attributes.DmgP] = bonus
			return m, true
		},
	})
}

// スキルヒット時の耐性低下
// また、Laumaの元素スキルまたは霜林の聖域の攻撃が敵に命中した時、
// その敵の草元素耐性と水元素耐性が10秒間低下する。
func (c *char) applyResReduction() {
	// ダメージイベントコールバック経由で耐性低下を適用
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		if len(args) < 3 {
			return false
		}

		enemy, ok := args[0].(combat.Target)
		if !ok {
			return false
		}

		atk, ok := args[1].(*combat.AttackEvent)
		if !ok {
			return false
		}

		dmg, ok := args[2].(float64)
		if !ok || dmg == 0 {
			return false
		}

		// Laumaのスキルまたは聖域からの攻撃か確認
		if atk.Info.ActorIndex != c.Index {
			return false
		}

		if atk.Info.AttackTag != attacks.AttackTagElementalArt {
			return false
		}

		// 草元素と水元素の耐性低下を適用
		if e, ok := enemy.(interface{ AddResistMod(combat.ResistMod) }); ok {
			e.AddResistMod(combat.ResistMod{
				Base:  modifier.NewBaseWithHitlag("lauma-dendro-res-reduction", 10*60),
				Ele:   attributes.Dendro,
				Value: -0.2,
			})
			e.AddResistMod(combat.ResistMod{
				Base:  modifier.NewBaseWithHitlag("lauma-hydro-res-reduction", 10*60),
				Ele:   attributes.Hydro,
				Value: -0.2,
			})
		}

		return false
	}, "lauma-res-reduction")
}
