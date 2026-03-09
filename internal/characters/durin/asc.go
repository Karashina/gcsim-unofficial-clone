package durin

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/reactions"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	// A1 ステートキー
	a1ResShredKey  = "durin-a1-res-shred"
	a1DarkDecayKey = "durin-a1-dark-decay"

	// A1 持続時間および値
	a1ResShredDuration = 6 * 60 // 6秒

	// A4 ステートキー
	a4PrimordialKey = "durin-a4-primordial"

	// A4 持続時間および値
	a4Duration         = 20 * 60 // 20秒
	a4MaxStacks        = 10
	a4DmgPercentPerAtk = 0.03 // 攻撃力100あたり3%
	a4MaxDmgPercent    = 0.75 // 追加ダメージ上限75%
)

// A1: 神算の光顕 (Light Manifest of the Divine Calculus)
// 白焔の龍: 燃焼、過負荷、炎元素拡散、炎元素結晶化反応後、
// または燃焼中の敵に炎/草元素ダメージを与えた後、炎元素耐性と対応する元素耐性を20%減少
// 暗蝕の龍: デュリンの蒸発と溶解ダメージが40%増加
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}

	// 白焔の龍の耐性デバフ用に元素反応イベントを購読
	c.Core.Events.Subscribe(event.OnBurning, c.a1WhiteFlameReactionCB(attributes.Pyro, attributes.Dendro), "durin-a1-burning")
	c.Core.Events.Subscribe(event.OnOverload, c.a1WhiteFlameReactionCB(attributes.Pyro, attributes.Electro), "durin-a1-overload")
	c.Core.Events.Subscribe(event.OnSwirlPyro, c.a1WhiteFlameReactionCB(attributes.Pyro, attributes.Anemo), "durin-a1-pyro-swirl")
	c.Core.Events.Subscribe(event.OnCrystallizePyro, c.a1WhiteFlameReactionCB(attributes.Pyro, attributes.Geo), "durin-a1-pyro-crystallize")

	// 燃焼中の敵への炎/草元素ダメージを購読
	c.Core.Events.Subscribe(event.OnEnemyDamage, c.a1WhiteFlameOnDamageCB, "durin-a1-burning-dmg")

	// 暗蝕の龍: 蒸発と溶解の反応補正
	c.a1DarkDecayReactMod()
}

func (c *char) a1WhiteFlameReactionCB(ele1, ele2 attributes.Element) func(args ...interface{}) bool {
	return func(args ...interface{}) bool {
		if !c.StatusIsActive(dragonWhiteFlameKey) {
			return false
		}

		target := args[0].(combat.Target)
		e, ok := target.(*enemy.Enemy)
		if !ok {
			return false
		}

		// 耐性デバフ量を計算（20%、Hexereiボーナス時35%）
		resShred := 0.20
		if c.hasHexereiBonus() {
			resShred = 0.35 // 20% * 1.75
		}

		// 両元素に耐性デバフを適用
		e.AddResistMod(combat.ResistMod{
			Base:  modifier.NewBaseWithHitlag(a1ResShredKey+"-"+ele1.String(), a1ResShredDuration),
			Ele:   ele1,
			Value: -resShred,
		})
		if ele1 != ele2 {
			e.AddResistMod(combat.ResistMod{
				Base:  modifier.NewBaseWithHitlag(a1ResShredKey+"-"+ele2.String(), a1ResShredDuration),
				Ele:   ele2,
				Value: -resShred,
			})
		}

		c.Core.Log.NewEvent("Durin A1 RES shred applied", glog.LogCharacterEvent, c.Index).
			Write("element1", ele1.String()).
			Write("element2", ele2.String()).
			Write("shred", resShred)

		return false
	}
}

func (c *char) a1WhiteFlameOnDamageCB(args ...interface{}) bool {
	if !c.StatusIsActive(dragonWhiteFlameKey) {
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

	// 耐性デバフ量を計算
	resShred := 0.20
	if c.hasHexereiBonus() {
		resShred = 0.35
	}

	// 炎元素耐性デバフと対応する元素の耐性デバフを適用
	e.AddResistMod(combat.ResistMod{
		Base:  modifier.NewBaseWithHitlag(a1ResShredKey+"-pyro", a1ResShredDuration),
		Ele:   attributes.Pyro,
		Value: -resShred,
	})
	e.AddResistMod(combat.ResistMod{
		Base:  modifier.NewBaseWithHitlag(a1ResShredKey+"-"+atk.Info.Element.String(), a1ResShredDuration),
		Ele:   atk.Info.Element,
		Value: -resShred,
	})

	return false
}

func (c *char) a1DarkDecayReactMod() {
	// 暗蝕の龍がアクティブ時に蒸発と溶解の反応補正を適用
	reactMod := 0.40
	if c.hasHexereiBonus() {
		reactMod = 0.70 // 40% * 1.75
	}

	c.AddReactBonusMod(character.ReactBonusMod{
		Base: modifier.NewBase(a1DarkDecayKey, -1),
		Amount: func(ai combat.AttackInfo) (float64, bool) {
			if !c.StatusIsActive(dragonDarkDecayKey) {
				return 0, false
			}
			if ai.ActorIndex != c.Index {
				return 0, false
			}
			if !ai.Amped {
				return 0, false
			}
			return reactMod, true
		},
	})
}

func (c *char) a4OnBurst() {
	if c.Base.Ascension < 4 {
		return
	}

	// 元素爆発使用時にスタックをリセット
	c.primordialFusionStacks = a4MaxStacks
	c.primordialFusionExpiry = c.Core.F + a4Duration
	c.AddStatus(a4PrimordialKey, a4Duration, true)

	c.Core.Log.NewEvent("Durin A4: Primordial Fusion stacks gained", glog.LogCharacterEvent, c.Index).
		Write("stacks", c.primordialFusionStacks)
}

func (c *char) a4DragonAttackCB(a combat.AttackCB) {
	if c.Base.Ascension < 4 {
		return
	}

	// スタックがあるか確認
	if c.primordialFusionStacks <= 0 || c.Core.F >= c.primordialFusionExpiry {
		return
	}

	// 1スタック消費（複数の敵にヒットしても1回の攻撃につき1回のみ）
	c.primordialFusionStacks--

	// ダメージボーナスを計算: 攻撃力100あたり3%、最大75%
	// 仕様ではai.Multに乗算する必要があるが、コールバック内のため、
	// 固定ダメージまたは同様の仕組みで適用する必要がある
	totalAtk := c.TotalAtk()
	dmgBonus := (totalAtk / 100.0) * a4DmgPercentPerAtk
	if dmgBonus > a4MaxDmgPercent {
		dmgBonus = a4MaxDmgPercent
	}

	c.Core.Log.NewEvent("Durin A4: Primordial Fusion consumed", glog.LogCharacterEvent, c.Index).
		Write("stacks_remaining", c.primordialFusionStacks).
		Write("atk", totalAtk).
		Write("dmg_bonus", dmgBonus)

	// 注: 実際のダメージ変更は攻撃情報のセットアップで処理される
	// コールバック内ではダメージを変更できないため。このコールバックはスタック消費用。
}

// A1暗蝕の龍チェック用の蒸発と溶解の反応タイプ
func init() {
	_ = reactions.Vaporize
	_ = reactions.Melt
}
