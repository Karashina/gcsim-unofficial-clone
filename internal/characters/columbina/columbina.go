package columbina

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
	core.RegisterCharFunc(keys.Columbina, NewChar)
}

type char struct {
	*tmpl.Character
	// 元素スキル用のGravityシステム
	gravity           int
	gravityLC         int // Lunar-ChargedからのGravity
	gravityLB         int // Lunar-BloomからのGravity
	gravityLCrs       int // Lunar-CrystallizeからのGravity
	gravityRippleSrc  int
	gravityRippleExp  int
	lunarDomainActive bool
	lunarDomainSrc    int
	lunacyStacks      int
	lunacySrc         int
	moonridgeDew      int
	moonridgeICD      int
	c4ICD             int
	c4DominantType    string // 4凸ボーナス用の支配的タイプを追跡
	// Gravity蓄積状態
	activeGravityType string
	// 固有天賦4の追跡
	a4LCSrc     int
	a4LCCount   int
	a4LBSrc     int
	a4LCrsSrc   int
	a4LCrsCount int
}

const (
	skillKey         = "columbina-skill"
	gravityRippleKey = "columbina-gravity-ripple"
	newMoonOmenKey   = "columbina-new-moon-omen"
	particleICDKey   = "columbina-particle-icd"
)

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.SkillCon = 3
	c.BurstCon = 5

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum

	// 状態を初期化
	c.gravityRippleSrc = -1
	c.lunacyStacks = 0
	c.moonridgeDew = 0
	c.moonridgeICD = 0
	c.c4ICD = 0

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	// パッシブを初期化
	c.a0Init()
	c.a1Init()
	// 命ノ星座を初期化
	if c.Base.Cons >= 1 {
		c.c1Init()
	}
	if c.Base.Cons >= 3 {
		c.c3c5Init()
	}
	if c.Base.Cons >= 6 {
		c.c6Init()
	}

	// Lunar反応イベントを購読してGravity蓄積を開始
	c.subscribeToLunarReactions()

	return nil
}

func (c *char) ActionStam(a action.Action, p map[string]int) float64 {
	// Moondew Cleanseは緑の露が1以上の時スタミナを消費しない
	if a == action.ActionCharge && c.Core.Player.Verdant.Count() >= 1 {
		return 0
	}
	return c.Character.ActionStam(a, p)
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 10
	}
	return c.Character.AnimationStartDelay(k)
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "gravity":
		return c.gravity, nil
	case "gravity-lc":
		return c.gravityLC, nil
	case "gravity-lb":
		return c.gravityLB, nil
	case "gravity-lcrs":
		return c.gravityLCrs, nil
	case "lunacy":
		return c.lunacyStacks, nil
	case "lunar-domain":
		if c.lunarDomainActive {
			return 1, nil
		}
		return 0, nil
	case "moonridge-dew":
		return c.moonridgeDew, nil
	}
	return c.Character.Condition(fields)
}

// getDominantLunarTypeは最もGravityが蓄積されたLunar反応タイプを返す
// 戻り値: Lunar-Chargedは"lc"、Lunar-Bloomは"lb"、Lunar-Crystallizeは"lcrs"
func (c *char) getDominantLunarType() string {
	if c.gravityLB == 0 && c.gravityLCrs == 0 && c.gravityLC == 0 {
		// チームメイトの元素で優先度を確認: 雷->"lc"、草->"lb"、岩->"lcrs"
		hasElectro, hasDendro, hasGeo := false, false, false
		for _, ch := range c.Core.Player.Chars() {
			if ch == nil || ch.Base.Key == c.Base.Key {
				continue // 自分自身とnilをスキップ
			}
			switch ch.Base.Element.String() {
			case "Electro":
				hasElectro = true
			case "Dendro":
				hasDendro = true
			case "Geo":
				hasGeo = true
			}
		}

		if hasElectro {
			return "lc"
		}
		if hasDendro {
			return "lb"
		}
		if hasGeo {
			return "lcrs"
		}

		// フォールバック: ランダム
		r := c.Core.Rand.Float64()
		if r < 0.33 {
			return "lc"
		} else if r < 0.66 {
			return "lb"
		} else {
			return "lcrs"
		}
	}
	if c.gravityLC >= c.gravityLB && c.gravityLC >= c.gravityLCrs {
		return "lc"
	}
	if c.gravityLB >= c.gravityLC && c.gravityLB >= c.gravityLCrs {
		return "lb"
	}
	return "lcrs"
}
