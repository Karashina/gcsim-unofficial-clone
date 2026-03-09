package chiori

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
	core.RegisterCharFunc(keys.Chiori, NewChar)
}

type char struct {
	*tmpl.Character

	// dolls
	skillSearchAoE   float64
	skillDoll        *ticker // 1st doll
	rockDoll         *ticker // 1凸の2体目の人形 / 構造物
	constructChecker *ticker

	// 固有天賦1の追跡
	a1Triggered   bool
	a1AttackCount int

	a4Buff []float64

	// 命ノ星座
	c1Active bool
	kinus    []*ticker
	c2Ticker *ticker

	c4AttackCount int
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = base.SkillDetails.BurstEnergyCost
	c.NormalHitNum = normalHitNum
	c.BurstCon = 5
	c.SkillCon = 3

	w.Character = &c
	return nil
}

func (c *char) Init() error {
	c.a1TapestrySetup()
	c.a4()

	c.skillSearchAoE = 12
	c.c1()
	c.c4()

	return nil
}

func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	// 固有天賦1のウィンドウがアクティブかつフィールド上かチェック
	if a == action.ActionSkill && c.StatusIsActive(a1WindowKey) {
		return true, action.NoFailure
	}
	return c.Character.ActionReady(a, p)
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	switch k {
	case model.AnimationXingqiuN0StartDelay:
		return 11
	case model.AnimationYelanN0StartDelay:
		return 3
	default:
		return c.Character.AnimationStartDelay(k)
	}
}
