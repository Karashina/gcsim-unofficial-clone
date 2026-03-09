package character

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
)

func (c *Character) CalcHealAmount(hi *info.HealInfo) (float64, float64) {
	var hp, bonus float64
	switch hi.Type {
	case info.HealTypeAbsolute:
		hp = hi.Src
	case info.HealTypePercent:
		hp = c.MaxHP() * hi.Src
	}
	bonus = 1 + c.HealBonus() + hi.Bonus
	return hp, bonus
}

func (c *Character) Heal(hi *info.HealInfo) (float64, float64) {
	hp, bonus := c.CalcHealAmount(hi)

	// ログ用に前回のHP関連値を保存
	prevHPRatio := c.CurrentHPRatio()
	prevHP := c.CurrentHP()
	prevHPDebt := c.CurrentHPDebt()

	// 元の回復量を計算
	healAmt := hp * bonus

	// HPとHP負債を変更
	heal := c.CharWrapper.ReceiveHeal(hi, healAmt)

	// 超過回復量を計算
	overheal := prevHP + heal - c.MaxHP()
	if overheal < 0 {
		overheal = 0
	}

	c.Core.Log.NewEvent(hi.Message, glog.LogHealEvent, c.Index).
		Write("previous_hp_ratio", prevHPRatio).
		Write("previous_hp", prevHP).
		Write("previous_hp_debt", prevHPDebt).
		Write("base amount", hp).
		Write("bonus", bonus).
		Write("final amount", healAmt).
		Write("received amount", heal).
		Write("overheal", overheal).
		Write("current_hp_ratio", c.CurrentHPRatio()).
		Write("current_hp", c.CurrentHP()).
		Write("current_hp_debt", c.CurrentHPDebt()).
		Write("max_hp", c.MaxHP())

	c.Core.Events.Emit(event.OnHeal, hi, c.Index, heal, overheal, healAmt)

	return heal, healAmt
}

func (c *Character) Drain(di *info.DrainInfo) float64 {
	prevHPRatio := c.CurrentHPRatio()
	prevHP := c.CurrentHP()
	c.ModifyHPByAmount(-di.Amount)

	c.Core.Log.NewEvent(di.Abil, glog.LogHurtEvent, di.ActorIndex).
		Write("previous_hp_ratio", prevHPRatio).
		Write("previous_hp", prevHP).
		Write("amount", di.Amount).
		Write("current_hp_ratio", c.CurrentHPRatio()).
		Write("current_hp", c.CurrentHP()).
		Write("max_hp", c.MaxHP())
	c.Core.Events.Emit(event.OnPlayerHPDrain, di)
	return di.Amount
}

func (c *Character) ReceiveHeal(hi *info.HealInfo, healAmt float64) float64 {
	// HP負債を考慮した実際の回復量を計算
	// TODO: 負債の清算と同じ回復で治癒が発生すると仮定している。次の回復からのみ発生する可能性もある
	// 例: HP負債が10、回復量が11の場合、回復なしではなく 11 - 10 = 1 だけ回復する
	heal := healAmt - c.CurrentHPDebt()
	if heal < 0 {
		heal = 0
	}

	// 元の回復量に基づきHP負債を更新
	c.ModifyHPDebtByAmount(-healAmt)

	// 実際の回復量に基づき回復を実行
	c.ModifyHPByAmount(heal)

	return heal
}
