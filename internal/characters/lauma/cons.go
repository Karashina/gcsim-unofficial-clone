package lauma

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	c1Key    = "lauma-c1-threads-of-life"
	c1ICDKey = "lauma-c1-heal-icd"
)

// 1凸
// Laumaが元素スキルまたは元素爆発を使用した後、20秒間「命の糸」を獲得する。
// この間、周囲のパーティメンバーがLunar-Bloom反応を発動した時、
// 周囲のアクティブキャラクターはLaumaの元素熟知の500%に等しいHPを回復する。この効果は1.9秒に1回発動可能。
func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}

	// 「命の糸」バフを20秒間付与
	c.AddStatus(c1Key, 20*60, true)

	// Lunar-Bloom反応時の回復を設定
	c.Core.Events.Subscribe(event.OnLunarBloom, func(args ...interface{}) bool {
		if !c.StatusIsActive(c1Key) {
			return false
		}
		if c.StatusIsActive(c1ICDKey) {
			return false
		}
		em := c.Stat(attributes.EM)
		healAmount := 5.0 * em // 500% of EM

		for _, char := range c.Core.Player.Chars() {
			char.Heal(&info.HealInfo{
				Caller:  c.Index,
				Target:  char.Index,
				Message: "Threads of Life",
				Src:     healAmount,
				Bonus:   0,
			})
		}

		// 1.9秒のICDを設定
		c.AddStatus(c1ICDKey, 114, true) // 1.9秒 × 60 = 114フレーム
		return false
	}, "lauma-c1-heal")
}

// 2凸
// 元素爆発発動時にムーンサイン：Ascendant Gleamがアクティブの場合、周囲のパーティメンバー全員のLunar-Bloomダメージが40%増加する。
func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}

	if c.MoonsignAscendant {
		// 元素爆発効果の持続時間中、Lunar-Bloomダメージ40%ボーナスを適用
		for _, char := range c.Core.Player.Chars() {
			char.AddLBReactBonusMod(character.LBReactBonusMod{
				Base: modifier.NewBase("lauma-c2-ascendant-lb-boost", 15*60), // 淡き讃歌と同じ持続時間
				Amount: func(ai combat.AttackInfo) (float64, bool) {
					return 0.4, false // 40%加算ボーナス
				},
			})
		}
	}
}

// 4凸
// 元素スキルで召喚した霜林の聖域の攻撃が敵に命中した時、
// Laumaは元素エネルギーを4回復する。この効果は5秒に1回発動可能。
func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}
	if c.StatusIsActive("lauma-c4-energy-icd") {
		return
	}
	c.AddEnergy("lauma C4", 4)
	c.AddStatus("lauma-c4-energy-icd", 5*60, true) // 5秒のICD
}

// 6凸
// 霜林の聖域が敵を攻撃した時、Laumaの元素熟知の185%に等しい範囲草元素ダメージを追加で1回与える。
// このダメージはLunar-Bloomダメージとみなされる。このダメージは淡き讃歌スタックを消費せず、Laumaに淡き讃歌を2スタック付与し、
// この方法で獲得した淡き讃歌スタックの持続時間を更新する。
// この効果は各霜林の聖域につき最大8回発生する。
// 元素スキル「ルーノ：カルシッコの夜明けなき安息」使用時、この方法で獲得した淡き讃歌スタックは全て削除される。
// また、Laumaが淡き讃歌スタックを持っている間に通常攻撃を使用すると、
// 1スタックを消費して元素熟知の150%に等しい草元素ダメージに変換する。このダメージはLunar-Bloomダメージとみなされる。
// ムーンサイン：Ascendant Gleam：周囲のパーティメンバー全員のLunar-Bloomダメージが1.25倍になる。

// 6凸：通常攻撃変換のヘルパー
func (c *char) c6NormalAttackConversion() bool {
	if c.Base.Cons < 6 {
		return false
	}
	if c.paleHymn <= 0 {
		return false
	}

	// 淡き讃歌を1スタック消費
	c.paleHymn--

	// 元素熟知の150%に等しい草元素ダメージをLunar-Bloomダメージとして与える
	em := c.Stat(attributes.EM)
	ai := combat.AttackInfo{
		ActorIndex:       c.Index,
		Abil:             "Normal Attack (C6 Conversion)",
		AttackTag:        attacks.AttackTagLBDamage,
		ICDTag:           attacks.ICDTagNone,
		StrikeType:       attacks.StrikeTypeDefault,
		Element:          attributes.Dendro,
		Durability:       0,
		IgnoreDefPercent: 1,
	}

	ai.FlatDmg = (1.5*em*(1+c.LBBaseReactBonus(ai)))*(1+((6*em)/(2000+em))+c.LBReactBonus(ai)) + c.burstLBBuff // 元素熟知の150%
	snap := combat.Snapshot{
		CharLvl: c.Base.Level,
	}
	snap.Stats[attributes.CR] = c.Stat(attributes.CR)
	snap.Stats[attributes.CD] = c.Stat(attributes.CD)

	c.Core.QueueAttackWithSnap(
		ai,
		snap,
		combat.NewSingleTargetHit(c.Core.Combat.PrimaryTarget().Key()),
		1,
	)
	return true
}

// 6凸 ムーンサイン：Ascendant Gleam - 周囲のパーティメンバー全員のLunar-Bloomダメージが1.25倍
// ElevationModで実装（LBダメージに+25%のElevation）
func (c *char) c6Init() {
	if c.Base.Cons < 6 {
		return
	}
	if !c.MoonsignAscendant {
		return
	}

	// 全パーティメンバーにLunar-Bloomダメージの25% Elevationボーナスを適用
	for _, char := range c.Core.Player.Chars() {
		char.AddElevationMod(character.ElevationMod{
			Base: modifier.NewBase("lauma-c6-ascendant-elevation", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				if ai.AttackTag == attacks.AttackTagLBDamage {
					return 0.25, false
				}
				return 0, false
			},
		})
	}
}
