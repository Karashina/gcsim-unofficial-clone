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
// ルミのLCrsダメージ時、1スタック消費して防御力75%分のダメージを追加。
// ミリオントンクラッシュでは最大5スタック消費し、各スタックあたり防御力150%分のダメージを追加。
func (c *char) c1Init() {
	// 元素スキル使用時にField Catalogスタック追加
	c.Core.Events.Subscribe(event.OnSkill, func(args ...interface{}) bool {
		if c.Core.Player.Active() != c.Index {
			return false
		}
		c.addFieldCatalogStacks(6)
		return false
	}, "linnea-c1-skill-stacks")

	// Lunar-Crystallize反応時にField Catalogスタック追加
	c.Core.Events.Subscribe(event.OnLunarCrystallize, func(args ...interface{}) bool {
		c.addFieldCatalogStacks(6)
		return false
	}, "linnea-c1-lcrs-stacks")

	c.Core.Log.NewEvent("Linnea C1 active: Field Catalog system initialized", glog.LogCharacterEvent, c.Index)
}

// c1OnSkillUse はスキル使用時のC1処理（summonLumiからの呼び出し用）
func (c *char) c1OnSkillUse() {
	c.addFieldCatalogStacks(6)
}

// c1OnMoondriftHarmony はMoondrift Harmony発動時のC1処理
func (c *char) c1OnMoondriftHarmony() {
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
		// C6: 2スタック消費で150%防御力ボーナス
		consume = 2
		bonusMult = 1.50
	}
	if c.fieldCatalogStacks < consume {
		consume = c.fieldCatalogStacks
		if c.Base.Cons >= 6 && consume < 2 {
			bonusMult = 0.75 // 1スタックしかない場合は通常倍率
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
		// C6: 2倍消費で各スタック150%追加
		maxConsume = 10
	}

	consume := min(maxConsume, c.fieldCatalogStacks)
	effectiveStacks := consume
	if c.Base.Cons >= 6 {
		effectiveStacks = consume / 2
		if effectiveStacks == 0 {
			effectiveStacks = 1
		}
		bonusPerStack = 1.50
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
	c.Core.Log.NewEvent("Linnea C2 active: Moondrift CRIT DMG bonus initialized", glog.LogCharacterEvent, c.Index)
}

// c2OnMoondriftHarmony はMoondrift Harmony発動時のC2処理
// 水元素または岩元素のパーティメンバーに会心ダメージ+40%を付与（8秒）
func (c *char) c2OnMoondriftHarmony() {
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

// c4OnMoondriftHarmony はMoondrift Harmony発動時のC4処理
func (c *char) c4OnMoondriftHarmony() {
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

	// アクティブキャラにもDEF+25%（リンネアがアクティブの場合は重複しない）
	activeIdx := c.Core.Player.Active()
	if activeIdx != c.Index {
		activeChar := c.Core.Player.ByIndex(activeIdx)
		activeChar.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(c4DefKey, c4Duration),
			AffectedStat: attributes.DEFP,
			Amount: func() ([]float64, bool) {
				return defMod, true
			},
		})
	}

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
			if ai.ActorIndex != c.Index {
				return 0, false
			}
			return 0.25, false
		},
	})

	c.Core.Log.NewEvent("Linnea C6 active: Enhanced Field Catalog and Elevation bonus", glog.LogCharacterEvent, c.Index)
}
