package clorinde

import (
	"strings"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
)

func (c *char) ReceiveHeal(hi *info.HealInfo, healAmt float64) float64 {
	// スキル状態中は回復無効。それ以外は通常通り
	if !c.StatusIsActive(skillStateKey) {
		return c.Character.ReceiveHeal(hi, healAmt)
	}

	// クロリンデ自身の回復はデフォルトで保持
	if hi.Caller == c.Index && strings.HasPrefix(hi.Message, "Impale the Night") {
		return c.Character.ReceiveHeal(hi, healAmt)
	}

	// 回復量を命の契約に変換
	factor := healingBOL[c.TalentLvlSkill()]
	if c.Base.Ascension >= 4 {
		factor = 1
	}

	amt := healAmt * factor
	c.Core.Log.NewEvent("clorinde healing surpressed", glog.LogHealEvent, c.Index).
		Write("bol_amount", amt)
	c.ModifyHPDebtByAmount(amt)

	return 0
}
