package linnea

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// c1Init は1凸を初期化する:
// 元素スキル使用時とLunar-Crystallize反応時に「Field Catalog」スタックを6獲得（最大18）。
// 近くのパーティメンバーのLCrsダメージ時、1スタック消費して防御力75%分のダメージを追加（LCrsFlatBonusMod）。
// ミリオントンクラッシュでは最大5スタック消費し、各スタックあたり防御力150%分のダメージを追加。
func (c *char) c1Init() {
	// Subscribe to base Moondrift Harmony: first moondrift projectile hit (AttackTagLCrsDamage + abil "lunar-crystallize")
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		ae, ok := args[1].(*combat.AttackEvent)
		if !ok {
			return false
		}
		if ae.Info.AttackTag != attacks.AttackTagLCrsDamage {
			return false
		}
		if ae.Info.Abil != "lunar-crystallize" {
			return false
		}
		c.onMoondriftHarmony()
		return false
	}, "linnea-moondrift-harmony")

	// Add LCrsFlatBonusMod to all party members.
	// When any party member (including Linnea herself) deals LCrs DMG,
	// consume 1 Field Catalog stack to add DEF×75% flat damage (C6: 2 stacks, DEF×112.5%).
	for _, char := range c.Core.Player.Chars() {
		char.AddLCrsFlatBonusMod(character.LCrsFlatBonusMod{
			Base: modifier.NewBase("linnea-c1-lcrs-flat", -1),
			Amount: func(atk combat.AttackInfo) (float64, bool) {
				return c.c1LCrsDamageBonus(), false
			},
		})
	}

	c.Core.Log.NewEvent("Linnea C1 active: Field Catalog system initialized", glog.LogCharacterEvent, c.Index)
}

// c1OnSkillUse はスキル使用時のC1処理（summonLumiからの呼び出し用）
func (c *char) c1OnSkillUse() {
	c.addFieldCatalogStacks(6)
}

// c1OnHarmony はMoondrift Harmony発動時のC1処理
func (c *char) c1OnHarmony() {
	c.addFieldCatalogStacks(6)
}

// addFieldCatalogStacks はField Catalogスタックを追加する
func (c *char) addFieldCatalogStacks(n int) {
	maxStacks := maxFieldCatalog
	if c.Base.Cons >= 6 {
		// C6: トリガー時に最大スタック数まで即座に追加
		c.fieldCatalogStacks = maxStacks
	} else {
		c.fieldCatalogStacks = min(maxStacks, c.fieldCatalogStacks+n)
	}
	c.fieldCatalogSrc = c.Core.F
	c.AddStatus(fieldCatalogKey, fieldCatalogDuration, true)

	c.Core.Log.NewEvent("Field Catalog stacks updated", glog.LogCharacterEvent, c.Index).
		Write("stacks", c.fieldCatalogStacks)
}

// c1LCrsDamageBonus はルミのLCrsダメージ時のField Catalog消費ボーナスを返す
// 1スタック消費して防御力75%分のダメージを追加
func (c *char) c1LCrsDamageBonus() float64 {
	if c.fieldCatalogStacks <= 0 || !c.StatusIsActive(fieldCatalogKey) {
		return 0
	}

	consume := 1
	bonusMult := 0.75
	if c.Base.Cons >= 6 {
		// C6: consume twice (2 stacks), DMG boosted to 150% of original (DEF*75%*1.5 = DEF*112.5%)
		consume = 2
		bonusMult = 1.125
	}
	if c.fieldCatalogStacks < consume {
		consume = c.fieldCatalogStacks
		if c.Base.Cons >= 6 && consume < 2 {
			bonusMult = 0.75 // not enough stacks for full C6: fall back to base rate
		}
	}

	c.fieldCatalogStacks -= consume
	bonus := c.TotalDef(false) * bonusMult

	c.Core.Log.NewEvent("Field Catalog consumed for LCrs damage", glog.LogCharacterEvent, c.Index).
		Write("consumed", consume).
		Write("bonus", bonus).
		Write("remaining", c.fieldCatalogStacks)

	return bonus
}

// c1MillionTonCrushBonus はミリオントンクラッシュのField Catalog消費ボーナスを返す
// 最大5スタック消費し、各スタックあたり防御力150%分のダメージを追加
func (c *char) c1MillionTonCrushBonus() float64 {
	if c.fieldCatalogStacks <= 0 || !c.StatusIsActive(fieldCatalogKey) {
		return 0
	}

	maxConsume := 5
	bonusPerStack := 1.50
	if c.Base.Cons >= 6 {
		// C6: consume twice (up to 10 stacks), DMG boosted to 150% of original (DEF*150%*1.5 = DEF*225% per pair)
		maxConsume = 10
		bonusPerStack = 2.25
	}

	consume := min(maxConsume, c.fieldCatalogStacks)
	effectiveStacks := consume
	if c.Base.Cons >= 6 {
		effectiveStacks = consume / 2
		if effectiveStacks == 0 {
			effectiveStacks = 1
		}
	}

	c.fieldCatalogStacks -= consume
	bonus := c.TotalDef(false) * bonusPerStack * float64(effectiveStacks)

	c.Core.Log.NewEvent("Field Catalog consumed for Million Ton Crush", glog.LogCharacterEvent, c.Index).
		Write("consumed", consume).
		Write("effectiveStacks", effectiveStacks).
		Write("bonus", bonus).
		Write("remaining", c.fieldCatalogStacks)

	return bonus
}

// c2Init は2凸を初期化する:
// Moondrift Harmony発動後、パーティ内の水/岩元素キャラの会心ダメージが40%増加（8秒）。
// ミリオントンクラッシュの会心ダメージが150%増加。
// Moonsign: Ascendant の場合、HOHとMTCもMoondrift Harmonyを発動する。
func (c *char) c2Init() {
	// C2 Ascendant Gleam: HOH/MTC trigger Moondrift Harmony (C1/C2/C4 effects)
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		if !c.MoonsignAscendant {
			return false
		}
		ae, ok := args[1].(*combat.AttackEvent)
		if !ok {
			return false
		}
		if ae.Info.ActorIndex != c.Index {
			return false
		}
		abil := ae.Info.Abil
		if abil != "Lumi Heavy Overdrive Hammer (Lunar-Crystallize)" &&
			abil != "Lumi Million Ton Crush (Lunar-Crystallize)" {
			return false
		}
		c.onMoondriftHarmony()
		return false
	}, "linnea-ascendant-harmony")

	c.Core.Log.NewEvent("Linnea C2 active: Moondrift CRIT DMG bonus initialized", glog.LogCharacterEvent, c.Index)
}

// c2OnHarmony はMoondrift Harmony発動時のC2処理
// 水元素または岩元素のパーティメンバーに会心ダメージ+40%を付与（8秒）
func (c *char) c2OnHarmony() {
	const c2Duration = 8 * 60

	for _, char := range c.Core.Player.Chars() {
		ele := char.Base.Element
		if ele != attributes.Hydro && ele != attributes.Geo {
			continue
		}

		idx := char.Index
		cdMod := make([]float64, attributes.EndStatType)
		cdMod[attributes.CD] = 0.40

		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(c2CritDmgKey, c2Duration),
			AffectedStat: attributes.CD,
			Amount: func() ([]float64, bool) {
				_ = idx // キャプチャ
				return cdMod, true
			},
		})
	}

	c.Core.Log.NewEvent("Linnea C2: Hydro/Geo party CRIT DMG +40%", glog.LogCharacterEvent, c.Index).
		Write("duration", c2Duration)
}

// c4Init は4凸を初期化する:
// Moondrift Harmony発動後、リンネアとアクティブキャラの防御力が25%増加（5秒）。
func (c *char) c4Init() {
	c.Core.Log.NewEvent("Linnea C4 active: DEF buff on Moondrift Harmony", glog.LogCharacterEvent, c.Index)
}

// c4OnHarmony はMoondrift Harmony発動時のC4処理
func (c *char) c4OnHarmony() {
	const c4Duration = 5 * 60

	defMod := make([]float64, attributes.EndStatType)
	defMod[attributes.DEFP] = 0.25

	// リンネア自身にDEF+25%
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag(c4DefKey, c4Duration),
		AffectedStat: attributes.DEFP,
		Amount: func() ([]float64, bool) {
			return defMod, true
		},
	})

	// アクティブキャラにもDEF+25%（リンネアがアクティブの場合は重ね掛け）
	activeIdx := c.Core.Player.Active()
	activeChar := c.Core.Player.ByIndex(activeIdx)
	activeDefMod := make([]float64, attributes.EndStatType)
	activeDefMod[attributes.DEFP] = 0.25
	activeChar.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag(c4DefActiveKey, c4Duration),
		AffectedStat: attributes.DEFP,
		Amount: func() ([]float64, bool) {
			return activeDefMod, true
		},
	})

	c.Core.Log.NewEvent("Linnea C4: DEF +25% applied", glog.LogCharacterEvent, c.Index).
		Write("duration", c4Duration)
}

// c6Init は6凸を初期化する:
// Field Catalogのスタック取得が強化され、トリガー時に最大スタックまで即座に追加。
// スタック消費が2倍になり、ボーナスが150%に増加。
// Moonsign: Ascendant の場合、LCrsダメージのElevationが25%増加。
func (c *char) c6Init() {
	// Ascendant Gleam時にLCrsダメージのElevation+25%
	c.AddElevationMod(character.ElevationMod{
		Base: modifier.NewBase("linnea-c6-elevation", -1),
		Amount: func(ai combat.AttackInfo) (float64, bool) {
			if !c.MoonsignAscendant {
				return 0, false
			}
			if ai.AttackTag != attacks.AttackTagLCrsDamage {
				return 0, false
			}
			// C6: applies to all nearby party members' LCrs DMG
			return 0.25, false
		},
	})

	c.Core.Log.NewEvent("Linnea C6 active: Enhanced Field Catalog and Elevation bonus", glog.LogCharacterEvent, c.Index)
}
