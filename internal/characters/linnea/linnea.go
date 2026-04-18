package linnea

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
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
	c4DefActiveKey       = "linnea-c4-def-active"
	lumiDuration         = 1560 // tE->終了: 1560f (26.0s from skill start)
	lumiSuperTickRate    = 141  // スーパーフォーム: PPP 2hit→次PPP 1hit = 120f + hit間隔 21f
	lumiSuperPPPToHOH    = 109  // PPP→HOH: PPP hit間(21f) + PPP hit2→HOH hit(88f) = 109f
	lumiSuperHOHToPPP    = 61   // HOH→PPP: HOH hit→次PPP hit1 = 61f
	lumiStandardTickRate = 321  // スタンダードフォーム: PPP 2hit→次PPP 1hit = 300f + hit間隔 21f
	// 初回ティックの遅延 (Eタップ→ルミ初撃: 108f, Q→ルミ和主動: 106f)
	lumiFirstTickFromE        = 108                                                   // tE→ルミ初撃
	lumiFirstTickFromQ        = 106                                                   // Q→D (Q発動後のルミ初撃)
	lumiStdFirstTickAfterMash = 243                                                   // mE→hitmark(111) + hitmark→PPP1(132)
	skillCD                   = 18 * 60                                               // 18秒
	burstCD                   = 15 * 60                                               // 15秒
	burstCDDelay              = 2                                                     // CT開始: 2f
	burstInitHealDelay        = 96                                                    // Q→回復: 96f
	burstContHealStart        = 158                                                   // Q→継続回復開始: 158f
	burstHealTickRate         = 60                                                    // 回復間隔: 60f (1秒)
	burstHealTicks            = 12                                                    // 回復回数: 12回
	burstHealDuration         = burstContHealStart + burstHealTicks*burstHealTickRate // 回復ステータス持続時間
	fieldCatalogDuration      = 10 * 60                                               // Field Catalog持続: 10秒
	maxFieldCatalog           = 18
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

// onMoondriftHarmony はムーンドリフト・ハーモニーの発動を処理する
// （C1/C2/C4の効果をトリガー）
func (c *char) onMoondriftHarmony() {
	// C1: Field Catalogスタックを追加
	if c.Base.Cons >= 1 {
		c.c1OnHarmony()
	}
	// C2: 水/岩パーティメンバーの会心ダメージ増加
	if c.Base.Cons >= 2 {
		c.c2OnHarmony()
	}
	// C4: 防御力増加
	if c.Base.Cons >= 4 {
		c.c4OnHarmony()
	}
}
