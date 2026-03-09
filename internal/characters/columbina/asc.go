package columbina

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	// 天賦0: Moonsign Benediction
	moonsignKey   = "moonsignKey"
	lbKeyStatus   = "LB-Key"
	lcKeyStatus   = "LC-Key"
	lcrsKeyStatus = "LCrs-Key"

	// 固有天賦1: Lunacy
	lunacyKey      = "lunacy"
	lunacyMaxStack = 3
	lunacyDur      = 10 * 60 // 10秒
	lunacyCRBonus  = 0.05    // 5% per stack

	// 固有天賦2: Law of the New Moon
	a4Key = "law-of-new-moon"
)

// 固有天賦0: 月印の恭福
// - Columbinaがパーティにいる時、"moonsign"と"lcrs-Key"ステータスを設定
// - 感電 → Lunar-Charged、開花 → Lunar-Bloom、水元素結晶 → Lunar-Crystallizeに変換
// - 基礎ダメージボーナス = HP上限1000ごとに0.2%、最大7%
func (c *char) a0Init() {
	for _, char := range c.Core.Player.Chars() {
		char.AddStatus(moonsignKey, -1, false)
		char.AddStatus(lbKeyStatus, -1, false)
		char.AddStatus(lcKeyStatus, -1, false)
		char.AddStatus(lcrsKeyStatus, -1, false)

		char.AddLCBaseReactBonusMod(character.LCBaseReactBonusMod{
			Base: modifier.NewBase("Moonlight, Lent Unto You (A0/LC)", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				maxval := 0.07
				val := min(maxval, c.MaxHP()/1000*0.002)
				return val, false
			},
		})
		char.AddLBBaseReactBonusMod(character.LBBaseReactBonusMod{
			Base: modifier.NewBase("Moonlight, Lent Unto You (A0/LB)", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				maxval := 0.07
				val := min(maxval, c.MaxHP()/1000*0.002)
				return val, false
			},
		})
		char.AddLCrsBaseReactBonusMod(character.LCrsBaseReactBonusMod{
			Base: modifier.NewBase("Moonlight, Lent Unto You (A0/LCrs)", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				maxval := 0.07
				val := min(maxval, c.MaxHP()/1000*0.002)
				return val, false
			},
		})
	}
}

// 固有天賦1: Lunacy - スタック数に基づく会心率ボーナス
func (c *char) a1Init() {
	if c.Base.Ascension < 1 {
		return
	}

	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase(lunacyKey+"-crit", -1),
		AffectedStat: attributes.CR,
		Amount: func() ([]float64, bool) {
			if c.lunacyStacks <= 0 {
				return nil, false
			}
			m := make([]float64, attributes.EndStatType)
			m[attributes.CR] = float64(c.lunacyStacks) * lunacyCRBonus
			return m, true
		},
	})
}

// a1OnGravityInterferenceはGravity Interferenceトリガー時にLunacyスタックを追加
func (c *char) a1OnGravityInterference() {
	if c.Base.Ascension < 1 {
		return
	}

	// スタックを追加（最大3）
	c.lunacyStacks++
	if c.lunacyStacks > lunacyMaxStack {
		c.lunacyStacks = lunacyMaxStack
	}

	// 持続時間を更新
	c.AddStatus(lunacyKey, lunacyDur, true)

	c.Core.Log.NewEvent("Lunacy stack gained", glog.LogCharacterEvent, c.Index).
		Write("stacks", c.lunacyStacks).
		Write("crit_bonus", float64(c.lunacyStacks)*lunacyCRBonus)

	// スタック減衰をスケジュール
	c.lunacySrc = c.Core.F
	c.Core.Tasks.Add(c.lunacyDecay(c.Core.F), lunacyDur)
}

// lunacyDecayは持続時間終了後にLunacyスタックを削除
func (c *char) lunacyDecay(src int) func() {
	return func() {
		if c.lunacySrc != src {
			return
		}
		c.lunacyStacks = 0
		c.Core.Log.NewEvent("Lunacy expired", glog.LogCharacterEvent, c.Index)
	}
}
