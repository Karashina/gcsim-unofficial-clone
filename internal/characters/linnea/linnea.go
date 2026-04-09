package linnea

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Linnea, NewChar)
}

// ルミの形態
type lumiForm int

const (
	lumiFormNone     lumiForm = iota
	lumiFormSuper             // スーパーパワーフォーム
	lumiFormUltimate          // アルティメットパワーフォーム
	lumiFormStandard          // スタンダードパワーフォーム
)

type char struct {
	*tmpl.Character
	// ルミの状態
	lumiActive   bool
	lumiSrc      int
	lumiForm     lumiForm
	lumiTickSrc  int
	lumiComboIdx int // スーパーパワーフォームのコンボ位置 (0,1=パンチ, 2=ハンマー)
	// 命ノ星座用
	fieldCatalogStacks int
	fieldCatalogSrc    int
	c2MoondriftSrc     int
}

const (
	lumiKey              = "linnea-lumi-active"
	burstHealKey         = "linnea-burst-heal"
	particleICDKey       = "linnea-particle-icd"
	fieldCatalogKey      = "linnea-field-catalog"
	c2CritDmgKey         = "linnea-c2-critdmg"
	c4DefKey             = "linnea-c4-def"
	a1GeoResKey          = "linnea-a1-geo-res"
	a1GeoResAscendKey    = "linnea-a1-geo-res-ascend"
	lumiDuration         = 25 * 60 // 25秒
	lumiSuperTickRate    = 2 * 60  // スーパーパワー: 2秒間隔
	lumiStandardTickRate = 3 * 60  // スタンダードパワー: 3秒間隔
	skillCD              = 18 * 60 // 18秒
	burstCD              = 15 * 60 // 15秒
	burstHealDuration    = 12 * 60 // 12秒
	burstHealTickRate    = 2 * 60  // 回復ティック: 2秒間隔
	fieldCatalogDuration = 10 * 60 // Field Catalog持続: 10秒
	maxFieldCatalog      = 18
)

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.BurstCon = 5
	c.SkillCon = 3

	// 状態を初期化
	c.lumiActive = false
	c.lumiSrc = -1
	c.lumiForm = lumiFormNone
	c.lumiTickSrc = -1
	c.lumiComboIdx = 0
	c.fieldCatalogStacks = 0

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	// A0パッシブを初期化（LCrsキー、月相レベル、DEFベースLCrsボーナス）
	c.a0Init()
	// A1パッシブを初期化（岩元素耐性ダウン）
	c.a1Init()
	// A4パッシブを初期化（元素熟知バフ）
	if c.Base.Ascension >= 4 {
		c.a4Init()
	}
	// 命ノ星座を初期化
	if c.Base.Cons >= 1 {
		c.c1Init()
	}
	if c.Base.Cons >= 2 {
		c.c2Init()
	}
	if c.Base.Cons >= 4 {
		c.c4Init()
	}
	if c.Base.Cons >= 6 {
		c.c6Init()
	}
	return nil
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	return c.Character.AnimationStartDelay(k)
}

// ActionReady は元素スキルのmash使用要件をチェックする
func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	if a == action.ActionSkill && c.lumiActive && p["mash"] == 1 {
		// mashはルミがスーパーパワーフォームの時のみ使用可能
		if c.lumiForm == lumiFormSuper {
			return true, action.NoFailure
		}
		return false, action.SkillCD
	}
	return c.Character.ActionReady(a, p)
}

// Condition はGCSLからキャラクター状態をクエリする
func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "lumi-active":
		if c.lumiActive {
			return 1, nil
		}
		return 0, nil
	case "lumi-form":
		return int(c.lumiForm), nil
	case "field-catalog":
		return c.fieldCatalogStacks, nil
	case "moonsign-ascendant":
		if c.MoonsignAscendant {
			return 1, nil
		}
		return 0, nil
	}
	return c.Character.Condition(fields)
}

// triggerMoondriftHarmony はムーンドリフト・ハーモニーの発動を処理する
// （C1/C2/C4の効果をトリガー）
func (c *char) triggerMoondriftHarmony() {
	// C1: Field Catalogスタックを追加
	if c.Base.Cons >= 1 {
		c.c1OnMoondriftHarmony()
	}
	// C2: 水/岩パーティメンバーの会心ダメージ増加
	if c.Base.Cons >= 2 {
		c.c2OnMoondriftHarmony()
	}
	// C4: 防御力増加
	if c.Base.Cons >= 4 {
		c.c4OnMoondriftHarmony()
	}
}
