package alhaitham

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	c1IcdKey    = "alhaitham-c1-icd"
	c2MaxStacks = 4
)

// 投影攻撃が敵に命中した時、共相・理式摹写のCDが1.2秒減少する。
// この効果は1秒ごとに1回のみ発動可能。
func (c *char) c1(a combat.AttackCB) {
	// 1凸がICD中なら無視
	if c.StatusIsActive(c1IcdKey) {
		return
	}
	c.ReduceActionCooldown(action.ActionSkill, 72) // 1.2秒短縮
	c.AddStatus(c1IcdKey, 60, true)                // 1s icd affected by hitlag
}

// アルハイゼムが琢光鏡を生成した時、元素熟知が8秒間50増加する。
// 最大4スタック。
// 各スタックの持続時間は独立してカウントされる。
// この効果は琢光鏡の最大数に達しても発動する。
func (c *char) c2(generated int) {
	m := make([]float64, attributes.EndStatType)
	m[attributes.EM] = 50
	for i := 0; i < generated; i++ {
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(c2ModName(c.c2Counter+1), 480), // 8s
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
		c.c2Counter = (c.c2Counter + 1) % c2MaxStacks // スタックは互いに独立しており、循環する
	}
}

func c2ModName(num int) string {
	return fmt.Sprintf("alhaitham-c2-%v-stack", num)
}

// 殊境・顕象結縛が発動された時、消費・生成された琢光鏡の数に応じて以下の効果が発動する:
// ・消費した鏡ごとに、他の周囲のパーティーメンバーの元素熟知が15秒間30増加する。
// ・生成した鏡ごとに、アルハイゼムの草元素ダメージバーナスが15秒間10%増加する。
// 上記効果の持続中に再度殊境・顕象結縛を使用すると、既存の持続時間はクリアされる
func (c *char) c4Loss(consumed int) {
	if consumed <= 0 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.EM] = 30.0 * float64(consumed)
	for i, char := range c.Core.Player.Chars() {
		// アルハイゼムをスキップ
		if i == c.Index {
			continue
		}
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("alhaitham-c4-loss", 900),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}
}

func (c *char) c4Gain(generated int) {
	if generated <= 0 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.DendroP] = 0.1 * float64(generated)
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag("alhaitham-c4-gain", 900),
		AffectedStat: attributes.DendroP,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})
}

// アルハイゼムが以下の効果を得る:
// ・殊境・顕象結縛発動の2秒後、消費した鏡の数に関係なく琢光鏡を3枚生成する。
//
// ・琢光鏡が最大数の時に琢光鏡を生成した場合、
// 会心率10%、会心ダメージ70%が6秒間増加する。
// この効果が持続中に再度発動した場合、残り持続時間が6秒延長される。
const c6key = "alhaitham-c6"

func (c *char) c6(generated int) {
	m := make([]float64, attributes.EndStatType)
	m[attributes.CR] = 0.1
	m[attributes.CD] = 0.7
	for i := 0; i < generated; i++ {
		if c.StatModIsActive(c6key) {
			c.ExtendStatus(c6key, 360)
			c.Core.Log.NewEvent("c6 buff extended", glog.LogCharacterEvent, c.Index).Write("c6 expiry on", c.StatusExpiry(c6key))
		} else {
			c.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag((c6key), 360), // 6s
				AffectedStat: attributes.CR,
				Amount: func() ([]float64, bool) {
					return m, true
				},
			})
		}
	}
}
