package arlecchino

import (
	"fmt"

	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Arlecchino, NewChar)
}

type char struct {
	*tmpl.Character
	skillDebt             float64
	skillDebtMax          float64
	initialDirectiveLevel int
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = base.SkillDetails.BurstEnergyCost
	c.NormalHitNum = normalHitNum
	c.NormalCon = 3
	c.BurstCon = 5

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.naBuff()
	c.passive()

	c.a1OnKill()
	c.a4()

	c.c2()
	return nil
}

func (c *char) NextQueueItemIsValid(k keys.Char, a action.Action, p map[string]int) error {
	lastAction := c.Character.Core.Player.LastAction
	if k != c.Base.Key && a != action.ActionSwap {
		return fmt.Errorf("%v: Tried to execute %v when not on field", c.Base.Key, a)
	}

	if lastAction.Type == action.ActionCharge && lastAction.Param["early_cancel"] > 0 {
		// 重撃はダッシュまたはジャンプでのみ早期キャンセル可能
		switch a {
		case action.ActionDash, action.ActionJump: // デフォルトブロックのエラーをスキップ
		default:
			return fmt.Errorf("%v: Cannot early cancel Charged Attack with %v", c.Base.Key, a)
		}
	}

	// 他の長柄武器キャラと異なり、通常攻撃なしで重撃を使用可能
	if a == action.ActionCharge {
		return nil
	}
	return c.Character.NextQueueItemIsValid(k, a, p)
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	switch k {
	case model.AnimationXingqiuN0StartDelay:
		return 15
	case model.AnimationYelanN0StartDelay:
		return 7
	default:
		return c.Character.AnimationStartDelay(k)
	}
}

func (c *char) ReceiveHeal(hi *info.HealInfo, healAmt float64) float64 {
	// 自身の回復以外の全ての回復を無視
	if hi.Caller == c.Index && hi.Message == balemoonRisingHealAbil {
		return c.Character.ReceiveHeal(hi, healAmt)
	}
	return 0
}
