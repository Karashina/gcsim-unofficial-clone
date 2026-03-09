package illuga

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Illuga, NewChar)
}

type char struct {
	*tmpl.Character
	// 元素爆発状態
	orioleSongActive        bool
	orioleSongSrc           int
	nightingaleSongStacks   int
	c2StackCounter          int // 2命用の消費スタック追跡
	geoConstructBonusStacks int // 岩元素構築物からの追加スタック追跡（最大15）
	// 月相状態
	moonsignAscendant bool
	// 固有天賦4のパーティ構成追跡
	a4HydroCount int
	a4GeoCount   int
}

const (
	orioleSongKey      = "illuga-oriole-song-active"
	lightkeeperOathKey = "illuga-lightkeeper-oath"
	particleICDKey     = "illuga-particle-icd"
)

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.BurstCon = 5
	c.SkillCon = 3

	// 状態を初期化
	c.orioleSongActive = false
	c.orioleSongSrc = -1
	c.nightingaleSongStacks = 0
	c.c2StackCounter = 0
	c.moonsignAscendant = false

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	// 固有天賦4を初期化（パーティ構成追跡）
	c.a4Init()

	// 命ノ星座を初期化
	c.consInit()

	// A0のmoonsignKeyを付与（月相レベル+1）
	c.AddStatus("moonsignKey", -1, true)

	// 月相状態の更新を購読
	c.updateMoonsignState()

	return nil
}

// updateMoonsignState はパーティ全体のフラグに基づいてイルーガの内部月相状態を更新する
func (c *char) updateMoonsignState() {
	c.moonsignAscendant = c.MoonsignAscendant
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 8
	}
	return c.Character.AnimationStartDelay(k)
}

// Condition はGCSLからキャラクター状態をクエリすることを可能にする
func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "oriole-song":
		if c.orioleSongActive {
			return 1, nil
		}
		return 0, nil
	case "nightingale-stacks":
		return c.nightingaleSongStacks, nil
	case "moonsign-ascendant":
		if c.moonsignAscendant {
			return 1, nil
		}
		return 0, nil
	}
	return c.Character.Condition(fields)
}

// isMoonsignAscendant は月相がAscendant Gleamかチェックする
// 初期化時に設定されたパーティ全体の月相ステータスをクエリする
func (c *char) isMoonsignAscendant() bool {
	return c.MoonsignAscendant
}
