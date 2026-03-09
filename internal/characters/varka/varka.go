package varka

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
)

func init() {
	core.RegisterCharFunc(keys.Varka, NewChar)
}

// ステータス/バフキー定数
const (
	sturmUndDrangKey  = "varka-sturm-und-drang"
	fwaCDKey          = "varka-fwa-cd"
	fwaChargeICDKey   = "varka-fwa-charge-icd"
	a1Key             = "varka-a1"
	a4Key             = "varka-a4-azure-fang"
	a4ICDPrefix       = "varka-a4-icd-"
	c1LyricalKey      = "varka-c1-lyrical"
	c4Key             = "varka-c4-buff"
	c6FWAWindowKey    = "varka-c6-fwa-window"
	c6AzureWindowKey  = "varka-c6-azure-window"
	particleICDKey    = "varka-particle-icd"
	normalHitNum      = 5
	sturmNormalHitNum = 5
)

type char struct {
	*tmpl.Character

	// Sturm und Drang 状態
	sturmActive        bool               // S&Dモードかどうか
	sturmSrc           int                // S&Dインスタンス追跡用ソースID
	otherElement       attributes.Element // パーティから決定された「他」元素
	hasOtherEle        bool               // パーティに炎/水/雷/氷がいるか
	savedNormalCounter int

	// 四風昇天
	fwaCharges       int // 現在のFWAチャージ数（最大2、C1で3）
	fwaMaxCharges    int // 最大チャージ数（常に2）
	fwaCDEndFrame    int // 次のFWAチャージが利用可能になるフレーム
	cdReductionCount int // S&D発動ごとの通常攻撃CD削減カウンター（最大15）
	cdReductionMax   int // CD削減最大数（基本15）

	// A1 パーティ構成
	anemoCount   int     // パーティ内の風元素キャラ数
	sameEleCount int     // 炎/水/雷/氷の同一元素最大数
	a1MultFactor float64 // 1.0, 1.4, or 2.2

	// A4 スタック
	a4Stacks int
	a4Expiry int

	// ヘクセライシステム
	isHexerei   bool
	hasHexBonus bool // パーティに2人以上のヘクセライキャラがいるか
}

func NewChar(s *core.Core, w *character.CharWrapper, p info.CharacterProfile) error {
	c := char{
		isHexerei:      true,
		fwaMaxCharges:  2,
		cdReductionMax: 15,
	}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.SkillCon = 3
	c.BurstCon = 5

	// nohex=1パラメータでヘクセライを無効化
	if nohex, ok := p.Params["nohex"]; ok && nohex == 1 {
		c.isHexerei = false
	}

	w.Character = &c
	return nil
}

func (c *char) Init() error {
	// パーティの元素構成を決定
	c.determineOtherElement()

	// A1倍率を決定
	c.determineA1Mult()

	// ヘクセライボーナスを確認
	c.checkHexereiBonus()

	// A1パッシブを初期化（攻撃力基準ダメージボーナス）
	c.a1Init()

	// A4パッシブを初期化（拡散イベント購読）
	if c.Base.Ascension >= 4 {
		c.a4Init()
	}

	// 命ノ星座を初期化
	// C1: 即時チャージ付与は enterSturmUndDrang で処理
	if c.Base.Cons >= 2 {
		c.c2Init()
	}
	if c.Base.Cons >= 4 {
		c.c4Init()
	}

	return nil
}

// ActionReady は S&DモードでのFWAを含むスキルの利用可能性を処理する
func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	if a == action.ActionSkill && c.sturmActive {
		// S&DモードではスキルがFour Winds' Ascensionになる
		c.updateFWACharges()
		if c.fwaCharges > 0 {
			return true, action.NoFailure
		}
		// C6: ウィンドウによりFWAチャージなしでスキル使用可能
		if c.Base.Cons >= 6 && (c.StatusIsActive(c6AzureWindowKey) || c.StatusIsActive(c6FWAWindowKey)) {
			return true, action.NoFailure
		}
		return false, action.SkillCD
	}
	return c.Character.ActionReady(a, p)
}

// updateFWACharges はCDが完了したチャージがあるか確認する。
// Core.Tasksベースの通常攻撃コールバックタイミングと合わせるためc.Core.F（絶対フレーム）を使用。
func (c *char) updateFWACharges() {
	for c.fwaCharges < c.fwaMaxCharges && c.Core.F >= c.fwaCDEndFrame {
		c.fwaCharges++
		if c.fwaCharges < c.fwaMaxCharges {
			c.fwaCDEndFrame += 11 * 60
		}
	}
}

// Condition はキャラクター状態のクエリに応答する
func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "hexerei":
		return c.isHexerei, nil
	case "sturm-und-drang":
		return c.sturmActive, nil
	case "fwa-charges":
		return c.fwaCharges, nil
	case "other-element":
		return c.otherElement.String(), nil
	case "a4-stacks":
		if c.Core.F >= c.a4Expiry {
			return 0, nil
		}
		return c.a4Stacks, nil
	}
	return c.Character.Condition(fields)
}

// determineOtherElement はパーティから最優先元素を探す
// 優先度: Pyro > Hydro > Electro > Cryo
func (c *char) determineOtherElement() {
	c.hasOtherEle = false
	priorityElements := []attributes.Element{
		attributes.Pyro,
		attributes.Hydro,
		attributes.Electro,
		attributes.Cryo,
	}
	for _, ele := range priorityElements {
		for _, char := range c.Core.Player.Chars() {
			if char.Base.Element == ele {
				c.otherElement = ele
				c.hasOtherEle = true
				return
			}
		}
	}
	// 該当元素がない場合は風元素をデフォルトにする
	c.otherElement = attributes.Anemo
}

// determineA1Mult はパーティ構成に基づいてA1パッシブ倍率を計算する
func (c *char) determineA1Mult() {
	if c.Base.Ascension < 1 {
		c.a1MultFactor = 1.0
		return
	}

	c.anemoCount = 0
	eleCounts := map[attributes.Element]int{}
	for _, char := range c.Core.Player.Chars() {
		switch char.Base.Element {
		case attributes.Anemo:
			c.anemoCount++
		case attributes.Pyro, attributes.Hydro, attributes.Electro, attributes.Cryo:
			eleCounts[char.Base.Element]++
		}
	}

	// 炎/水/雷/氷の同一元素最大数を検索
	c.sameEleCount = 0
	for _, cnt := range eleCounts {
		if cnt > c.sameEleCount {
			c.sameEleCount = cnt
		}
	}

	// 倍率を決定:
	// 風2人以上かつ同一他元素2人以上: 2.2倍
	// 風2人以上または同一他元素2人以上: 1.4倍
	// その他: 1.0倍
	hasAnemoBonus := c.anemoCount >= 2
	hasOtherBonus := c.sameEleCount >= 2

	if hasAnemoBonus && hasOtherBonus {
		c.a1MultFactor = 2.2
	} else if hasAnemoBonus || hasOtherBonus {
		c.a1MultFactor = 1.4
	} else {
		c.a1MultFactor = 1.0
	}
}

// checkHexereiBonus はパーティに2人以上のヘクセライキャラがいるか確認する
func (c *char) checkHexereiBonus() {
	if !c.isHexerei {
		c.hasHexBonus = false
		return
	}
	hexereiCount := 0
	for _, char := range c.Core.Player.Chars() {
		if result, err := char.Condition([]string{"hexerei"}); err == nil {
			if isHex, ok := result.(bool); ok && isHex {
				hexereiCount++
			}
		}
	}
	c.hasHexBonus = hexereiCount >= 2
}

// getCDReductionAmount はヘクセライボーナスに基づく通常攻撃ヒットごとのCD削減量を返す
func (c *char) getCDReductionAmount() int {
	if c.hasHexBonus {
		return 60 // 1.0s with Hexerei Secret Rite
	}
	return 30 // 0.5s base
}
