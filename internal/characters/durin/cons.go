package durin

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	// 1凸ステートキー
	c1CycleKey = "durin-c1-cycle"

	// 1凸の持続時間および値
	c1CycleDuration          = 20 * 60 // 20秒
	c1MaxStacks              = 20
	c1PurityFlatDmgPercent   = 0.60 // デュリンの攻撃力の60%
	c1DarknessFlatDmgPercent = 1.50 // デュリンの攻撃力の150%
	c1DarknessStackCost      = 2    // 元素爆発ヒットごとに2スタック消費
	c1NoConsumeChance        = 0.30 // 4凸: 30%の確率でスタックを消費しない

	// 2凸ステートキー
	c2BuffKey    = "durin-c2-buff"
	c2EleBuffKey = "durin-c2-ele-buff"

	// 2凸の持続時間および値
	c2BuffDuration = 20 * 60 // 20秒
	c2EleDuration  = 6 * 60  // 6秒
	c2DmgBonus     = 0.50    // 50%ダメージボーナス

	// 4凸の値
	c4BurstDmgBonus = 0.40 // 40%元素爆発ダメージボーナス

	// 6凸ステートキー
	c6DefShredKey = "durin-c6-def-shred"

	// 6凸の持続時間および値
	c6DefShredDuration = 6 * 60 // 6秒
	c6DefShred         = 0.30   // 30% DEFデバフ
)

// 1凸: アダマの救済 (Adamah's Redemption)
// 元素爆発発動後:
// 純化: 他のパーティメンバーに啓示のサイクル20スタックを付与（20秒）
//
//	通常攻撃/重撃/落下攻撃/元素スキル/元素爆発ダメージを与えた時、1スタック消費してデュリンの攻撃力の60%を固定ダメージとして追加
//
// 暗闇: デュリンに啓示のサイクル20スタックを付与（20秒）
//
//	元素爆発ダメージを与えた時、2スタック消費してデュリンの攻撃力の150%を固定ダメージとして追加
func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}

	// 固定ダメージ適用のためOnEnemyHitを購読
	c.Core.Events.Subscribe(event.OnEnemyHit, c.c1OnEnemyHit, "durin-c1-flat-dmg")
}

func (c *char) c1OnBurstPurity() {
	if c.Base.Cons < 1 {
		return
	}

	// 他の全パーティメンバーにスタックを付与
	for _, char := range c.Core.Player.Chars() {
		if char.Index == c.Index {
			continue // デュリンをスキップ
		}
		c.cycleStacks[char.Index] = c1MaxStacks
		c.cycleExpiry[char.Index] = c.Core.F + c1CycleDuration
	}

	c.Core.Log.NewEvent("C1: Cycle of Enlightenment granted to party (Purity)", glog.LogCharacterEvent, c.Index).
		Write("stacks", c1MaxStacks)
}

func (c *char) c1OnBurstDarkness() {
	if c.Base.Cons < 1 {
		return
	}

	// デュリンのみにスタックを付与
	c.cycleStacks[c.Index] = c1MaxStacks
	c.cycleExpiry[c.Index] = c.Core.F + c1CycleDuration

	c.Core.Log.NewEvent("C1: Cycle of Enlightenment granted to Durin (Darkness)", glog.LogCharacterEvent, c.Index).
		Write("stacks", c1MaxStacks)
}

func (c *char) c1OnEnemyHit(args ...interface{}) bool {
	if c.Base.Cons < 1 {
		return false
	}

	atk := args[1].(*combat.AttackEvent)
	charIndex := atk.Info.ActorIndex

	// このキャラクターにスタックがあるか確認
	expiry, ok := c.cycleExpiry[charIndex]
	if !ok || c.Core.F >= expiry || c.cycleStacks[charIndex] <= 0 {
		return false
	}

	// デュリンの暗闇モードスタックか確認（元素爆発ダメージにのみ適用）
	if charIndex == c.Index {
		// デュリンのスタックは元素爆発ダメージにのみ有効（暗闇モード）
		if atk.Info.AttackTag != attacks.AttackTagElementalBurst {
			return false
		}

		// 2スタック消費
		stacksToConsume := c1DarknessStackCost

		// 4凸: 30%の確率でスタックを消費しない
		if c.Base.Cons >= 4 && c.Core.Rand.Float64() < c1NoConsumeChance {
			stacksToConsume = 0
			c.Core.Log.NewEvent("C4: Cycle of Enlightenment stacks not consumed", glog.LogCharacterEvent, c.Index)
		}

		if c.cycleStacks[charIndex] < stacksToConsume {
			stacksToConsume = c.cycleStacks[charIndex]
		}
		c.cycleStacks[charIndex] -= stacksToConsume

		// 固定ダメージを追加: デュリンの攻撃力の150%
		flatDmg := c.TotalAtk() * c1DarknessFlatDmgPercent
		atk.Info.FlatDmg += flatDmg

		c.Core.Log.NewEvent("C1: Cycle of Enlightenment (Darkness) flat DMG added", glog.LogCharacterEvent, c.Index).
			Write("flat_dmg", flatDmg).
			Write("stacks_remaining", c.cycleStacks[charIndex])

		return false
	}

	// 他のパーティメンバーのスタック（純化モード）
	// アクティブキャラクターか確認
	if charIndex != c.Core.Player.Active() {
		return false
	}

	// 有効な攻撃タイプか確認（通常攻撃、重撃、落下攻撃、元素スキル、元素爆発）
	validTag := atk.Info.AttackTag == attacks.AttackTagNormal ||
		atk.Info.AttackTag == attacks.AttackTagExtra ||
		atk.Info.AttackTag == attacks.AttackTagPlunge ||
		atk.Info.AttackTag == attacks.AttackTagElementalArt ||
		atk.Info.AttackTag == attacks.AttackTagElementalArtHold ||
		atk.Info.AttackTag == attacks.AttackTagElementalBurst

	if !validTag {
		return false
	}

	// 1スタック消費
	stacksToConsume := 1

	// 4凸: 30%の確率でスタックを消費しない
	if c.Base.Cons >= 4 && c.Core.Rand.Float64() < c1NoConsumeChance {
		stacksToConsume = 0
		c.Core.Log.NewEvent("C4: Cycle of Enlightenment stacks not consumed", glog.LogCharacterEvent, charIndex)
	}

	c.cycleStacks[charIndex] -= stacksToConsume

	// 固定ダメージを追加: デュリンの攻撃力の60%
	flatDmg := c.TotalAtk() * c1PurityFlatDmgPercent
	atk.Info.FlatDmg += flatDmg

	c.Core.Log.NewEvent("C1: Cycle of Enlightenment (Purity) flat DMG added", glog.LogCharacterEvent, charIndex).
		Write("flat_dmg", flatDmg).
		Write("stacks_remaining", c.cycleStacks[charIndex])

	return false
}

// 2凸: 不穏な幻視 (Unsound Visions)
// デュリンの元素爆発使用後20秒間、パーティメンバーが
// 蒸発、溶解、燃焼、過負荷、炎元素拡散、炎元素結晶化を発動した後、
// または燃焼中の敵に炎/草元素ダメージを与えた後、全パーティメンバーに
// 炎元素ダメージ50%と対応する元素ダメージボーナス50%を6秒間付与
func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}

	// 元素反応イベントを購読
	c.Core.Events.Subscribe(event.OnVaporize, c.c2ReactionCB(attributes.Hydro), "durin-c2-vaporize")
	c.Core.Events.Subscribe(event.OnMelt, c.c2ReactionCB(attributes.Cryo), "durin-c2-melt")
	c.Core.Events.Subscribe(event.OnBurning, c.c2ReactionCB(attributes.Dendro), "durin-c2-burning")
	c.Core.Events.Subscribe(event.OnOverload, c.c2ReactionCB(attributes.Electro), "durin-c2-overload")
	c.Core.Events.Subscribe(event.OnSwirlPyro, c.c2ReactionCB(attributes.Anemo), "durin-c2-pyro-swirl")
	c.Core.Events.Subscribe(event.OnCrystallizePyro, c.c2ReactionCB(attributes.Geo), "durin-c2-pyro-crystallize")

	// 燃焼中の敵への炎/草元素ダメージを購読
	c.Core.Events.Subscribe(event.OnEnemyDamage, c.c2OnDamageCB, "durin-c2-burning-dmg")
}

func (c *char) c2ReactionCB(otherEle attributes.Element) func(args ...interface{}) bool {
	return func(args ...interface{}) bool {
		// 元素爆発後20秒以内か確認
		if !c.StatusIsActive(c2BuffKey) {
			return false
		}

		c.c2ApplyBuff(otherEle)
		return false
	}
}

func (c *char) c2OnDamageCB(args ...interface{}) bool {
	if !c.StatusIsActive(c2BuffKey) {
		return false
	}

	atk := args[1].(*combat.AttackEvent)
	target := args[0].(combat.Target)
	e, ok := target.(*enemy.Enemy)
	if !ok {
		return false
	}

	// 対象が燃焼中か確認
	if !e.AuraContains(attributes.Pyro, attributes.Dendro) {
		return false
	}

	// 炎または草元素ダメージを与えているか確認
	if atk.Info.Element != attributes.Pyro && atk.Info.Element != attributes.Dendro {
		return false
	}

	c.c2ApplyBuff(atk.Info.Element)
	return false
}

func (c *char) c2ApplyBuff(otherEle attributes.Element) {
	// 全パーティメンバーに炎元素ダメージ50%と対応する元素ダメージボーナスを適用
	for _, char := range c.Core.Player.Chars() {
		// 炎元素ダメージボーナス
		m := make([]float64, attributes.EndStatType)
		m[attributes.PyroP] = c2DmgBonus
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(c2EleBuffKey+"-pyro", c2EleDuration),
			AffectedStat: attributes.PyroP,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})

		// 対応する元素ダメージボーナス
		if otherEle != attributes.Pyro {
			m2 := make([]float64, attributes.EndStatType)
			switch otherEle {
			case attributes.Hydro:
				m2[attributes.HydroP] = c2DmgBonus
			case attributes.Cryo:
				m2[attributes.CryoP] = c2DmgBonus
			case attributes.Electro:
				m2[attributes.ElectroP] = c2DmgBonus
			case attributes.Anemo:
				m2[attributes.AnemoP] = c2DmgBonus
			case attributes.Geo:
				m2[attributes.GeoP] = c2DmgBonus
			case attributes.Dendro:
				m2[attributes.DendroP] = c2DmgBonus
			}
			char.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag(c2EleBuffKey+"-"+otherEle.String(), c2EleDuration),
				AffectedStat: attributes.NoStat,
				Amount: func() ([]float64, bool) {
					return m2, true
				},
			})
		}
	}

	c.Core.Log.NewEvent("C2: Elemental DMG bonus applied to party", glog.LogCharacterEvent, c.Index).
		Write("pyro_bonus", c2DmgBonus).
		Write("other_element", otherEle.String())
}

// 4凸: エマナレの源泉 (Emanare's Source)
// デュリンの元素爆発ダメージが40%増加
// 追加で、啓示のサイクルのスタックを消費しない確犘30%（c1OnEnemyHitで処理）
func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}

	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = c4BurstDmgBonus

	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("durin-c4-burst-dmg", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagElementalBurst {
				return nil, false
			}
			return m, true
		},
	})
}

// 6凸: 双つの誕生 (Dual Birth)
// 純化の真理: 元素爆発ダメージが30% DEF無視、白焔の龍がヒット時に敵のDEFを30%減少（6秒）
// 暗闇の真理: 元素爆発ダメージが70% DEF無視（基本30% + 追加40%）
// 注: 純化は「DEF無視 + DEFデバフ」の組み合わせ、暗闇は「高倍率DEF無視」のみ
func (c *char) c6DragonWhiteFlameCB(a combat.AttackCB) {
	if c.Base.Cons < 6 {
		return
	}

	e, ok := a.Target.(*enemy.Enemy)
	if !ok {
		return
	}

	// DEFデバフを適用
	e.AddDefMod(combat.DefMod{
		Base:  modifier.NewBaseWithHitlag(c6DefShredKey, c6DefShredDuration),
		Value: -c6DefShred,
	})

	c.Core.Log.NewEvent("C6: Dragon of White Flame DEF shred applied", glog.LogCharacterEvent, c.Index).
		Write("def_shred", c6DefShred)
}
