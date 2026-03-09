package columbina

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var burstFrames []int

const (
	burstHitmark      = 111
	lunarDomainDur    = 1288
	lunarDomainKey    = "lunar-domain"
	lunarDomainModKey = "lunar-domain-bonus"
)

func init() {
	burstFrames = frames.InitAbilSlice(125)
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	// 初期爆発ダメージ（AoE水元素ダメージ）
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Moonlit Melancholy (Q)",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 50,
	}
	ai.FlatDmg = c.MaxHP() * burstDmg[c.TalentLvlBurst()]

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 8)
	c.Core.QueueAttack(ai, ap, burstHitmark, burstHitmark)

	// Lunar Domainを有効化
	c.AddStatus(lunarDomainKey, lunarDomainDur, true)
	c.lunarDomainSrc = c.Core.F
	c.lunarDomainActive = true

	// 全パーティメンバーにLunar Domainバフを適用
	c.applyLunarDomainBuff()

	c.Core.Log.NewEvent("Lunar Domain activated", glog.LogCharacterEvent, c.Index).
		Write("duration", lunarDomainDur).
		Write("bonus", burstBonus[c.TalentLvlBurst()])

	// クリーンアップをスケジュール
	c.Core.Tasks.Add(func() {
		if c.lunarDomainSrc != c.Core.F-lunarDomainDur {
			return
		}
		c.lunarDomainActive = false
	}, lunarDomainDur)

	// エネルギーとクールダウン
	c.ConsumeEnergy(3)
	c.SetCDWithDelay(action.ActionBurst, 15*60, 1)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash],
		State:           action.BurstState,
	}, nil
}

// applyLunarDomainBuffは全パーティメンバーにLunar反応ダメージボーナスを適用
func (c *char) applyLunarDomainBuff() {
	if c.Base.Ascension >= 4 {
		for _, char := range c.Core.Player.Chars() {
			char.AddStatus(a4Key, lunarDomainDur, false)
		}
	}

	bonus := burstBonus[c.TalentLvlBurst()]
	dur := lunarDomainDur

	for _, char := range c.Core.Player.Chars() {
		// Lunar-Charged反応ボーナスを追加
		char.AddLCReactBonusMod(character.LCReactBonusMod{
			Base: modifier.NewBaseWithHitlag(lunarDomainModKey+"-lc", dur),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				return bonus, false
			},
		})

		// Lunar-Bloom反応ボーナスを追加
		char.AddLBReactBonusMod(character.LBReactBonusMod{
			Base: modifier.NewBaseWithHitlag(lunarDomainModKey+"-lb", dur),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				return bonus, false
			},
		})

		// Lunar-Crystallize反応ボーナスを追加
		char.AddLCrsReactBonusMod(character.LCrsReactBonusMod{
			Base: modifier.NewBaseWithHitlag(lunarDomainModKey+"-lcrs", dur),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				return bonus, false
			},
		})
	}
}
