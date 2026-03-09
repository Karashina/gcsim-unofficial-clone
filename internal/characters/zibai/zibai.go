package zibai

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
	core.RegisterCharFunc(keys.Zibai, NewChar)
}

type char struct {
	*tmpl.Character
	// 月相転移の状態
	lunarPhaseShiftActive bool
	lunarPhaseShiftSrc    int
	phaseShiftRadiance    int // 現在の輝度ポイント
	maxPhaseShiftRadiance int // デフォルト300
	spiritSteedUsages     int // モードごとの現在の使用回数
	maxSpiritSteedUsages  int // デフォルト3、1凸で5に増加
	savedNormalCounter    int // 4凸用の保存された通常攻撃カウンター
	// 命ノ星座追跡用
	c1FirstStride     bool // 1凸: 初回ストライドボーナス
	c4ScattermoonUsed bool // 4凸: 次のN4が250%ダメージ

}

const (
	skillKey                 = "zibai-lunar-phase-shift"
	selenicDescentKey        = "zibai-selenic-descent"
	spiritSteedCDKey         = "zibai-spirit-steed-cd"
	particleICDKey           = "zibai-particle-icd"
	radianceNormalICDKey     = "zibai-radiance-na-icd"
	radianceLCrsICDKey       = "zibai-radiance-lcrs-icd"
	lunarPhaseShiftDuration  = 990 // 16.5 seconds
	spiritSteedRadianceCost  = 70  // 神馬駆けの使用コスト
	normalPhaseShiftRadiance = 100 // デフォルト最大輝度（仕様: 最大100まで）
	c6ElevationBuffKey       = "zibai-c6-elevation"
	radianceTickInterval     = 6
	radianceTickGain         = 1
	radianceNormalGain       = 5
	radianceNormalICD        = 30
	radianceLCrsGain         = 35
	radianceLCrsICD          = 4 * 60
)

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.BurstCon = 5
	c.SkillCon = 3
	// 状態を初期化
	c.lunarPhaseShiftActive = false
	c.phaseShiftRadiance = 0
	c.maxPhaseShiftRadiance = normalPhaseShiftRadiance
	c.maxSpiritSteedUsages = 4
	c.spiritSteedUsages = 0
	c.c1FirstStride = false
	c.c4ScattermoonUsed = false

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	// 月相パッシブを初期化（パーティにLCrsキーを付与）
	c.a0Init()
	// 固有天賦1を初期化
	c.a1Init()
	// 固有天賦2を初期化
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

	c.initRadianceHandlers()

	return nil
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 10
	}
	return c.Character.AnimationStartDelay(k)
}

// Condition はGCSLからキャラクターの状態を問い合わせる
func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "lunar-phase-shift":
		if c.lunarPhaseShiftActive {
			return 1, nil
		}
		return 0, nil
	case "phase-shift-radiance":
		return c.phaseShiftRadiance, nil
	case "spirit-steed-usages":
		return c.spiritSteedUsages, nil
	case "moonsign-ascendant":
		if c.MoonsignAscendant {
			return 1, nil
		}
		return 0, nil
	}
	return c.Character.Condition(fields)
}

// ActionReady は神馬駆けが使用可能かチェックする
func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	// Spirit Steed's Strideの特殊チェック
	if a == action.ActionSkill && c.lunarPhaseShiftActive {
		// 再発動可能かチェック（輝度と使用回数）
		if c.phaseShiftRadiance < spiritSteedRadianceCost {
			return false, action.InsufficientEnergy
		}
		if c.spiritSteedUsages >= c.maxSpiritSteedUsages {
			return false, action.SkillCD
		}
		return true, action.NoFailure
	}
	return c.Character.ActionReady(a, p)
}

// isMoonsignAscendant は月相がAscendant Gleamかチェックする
func (c *char) isMoonsignAscendant() bool {
	return c.MoonsignAscendant
}

// addPhaseShiftRadiance は輝度ポイントを追加する（上限あり）
func (c *char) addPhaseShiftRadiance(amount int) {
	// 6凸: 獲得率50%増加
	if c.Base.Cons >= 6 {
		amount = int(float64(amount) * 1.5)
	}
	c.phaseShiftRadiance += amount
	if c.phaseShiftRadiance > c.maxPhaseShiftRadiance {
		c.phaseShiftRadiance = c.maxPhaseShiftRadiance
	}
}
