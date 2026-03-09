package dehya

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
)

const (
	a1ReductionKey      = "dehya-a1-reduction"
	a1ReductionDuration = 6 * 60
	a1ReductionMult     = 0.6
	a1ICDKey            = "dehya-a1-icd"
	a1ICD               = 2 * 60
	a4ICDKey            = "dehya-a4-icd"
	a4ICD               = 20 * 60
	a4HealMsg           = "Stalwart and True (A4)"
	a4Threshold         = 0.4
	a4InitialHealRatio  = 0.2
	a4DoTHealRatio      = 0.06
	a4DoTHealInterval   = 2 * 60
)

// ディヘヤが熔鉄の炎・烈炎潜伝または獅子拳で浄焔の領域を回収した後6秒間、
// 赤鬣の血のダメージを受ける際に受けるダメージが60%軽減される。
// この効果は2秒毎に1回発動可能。
func (c *char) a1Reduction() {
	if c.Base.Ascension < 1 {
		return
	}
	if c.StatusIsActive(a1ICDKey) {
		return
	}
	c.AddStatus(a1ICDKey, a1ICD, true)
	c.AddStatus(a1ReductionKey, a1ReductionDuration, true)
}

// TODO: 固有天賦1の中断耐性部分は未実装
// また、ディヘヤが熔鉄の炎・不屈の炎を発動した後9秒間、
// パーティ全体に黄金鍛造の姿状態を付与する。
// この状態は浄焔の領域内にいるキャラクターの中断耐性をさらに上昇させる。
// 黄金鍛造の姿は18秒毎に1回発動可能。

// HPが40%未満の際、ディヘヤはHP上限の20%分のHPを回復し、
// 以降10秒間、2秒毎にHP上限の6%分のHPを回復する。
// この効果は20秒毎に1回発動可能。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	// TODO: 本来は1秒毎にチェックすべきだが、これで十分…
	c.Core.Events.Subscribe(event.OnPlayerHPDrain, func(args ...interface{}) bool {
		di := args[0].(*info.DrainInfo)
		if di.Amount <= 0 {
			return false
		}
		if c.CurrentHPRatio() >= a4Threshold {
			return false
		}
		if c.StatusIsActive(a4ICDKey) {
			return false
		}
		c.AddStatus(a4ICDKey, a4ICD, true)
		// HP20%回復部分
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index,
			Target:  c.Index,
			Message: a4HealMsg,
			Src:     a4InitialHealRatio * c.MaxHP(),
			Bonus:   c.Stat(attributes.Heal),
		})
		// 10秒間2秒毎にHP6%回復部分（5回）
		c.QueueCharTask(c.a4DotHeal(0), a4DoTHealInterval)
		return false
	}, "hutao-c6")
}

// 複数のA4が同時に発動することは不可能なため、ソースは不要
func (c *char) a4DotHeal(count int) func() {
	return func() {
		if count == 5 {
			return
		}
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index,
			Target:  c.Index,
			Message: a4HealMsg,
			Src:     a4DoTHealRatio * c.MaxHP(),
			Bonus:   c.Stat(attributes.Heal),
		})
		c.QueueCharTask(c.a4DotHeal(count+1), a4DoTHealInterval)
	}
}
