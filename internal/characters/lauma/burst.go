package lauma

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
)

var burstFrames []int

func init() {
	burstFrames = frames.InitAbilSlice(116)
}

// 元素爆発
// 淡き讃歌を18スタック獲得する。
// また、Laumaが月の歌を持っている間に元素爆発を使用した場合、または元素爆発使用後15秒以内に月の歌を獲得した場合、
// 全ての月の歌スタックを消費し、消費した月の歌1スタックにつき淡き讃歌を6スタック獲得する。
// この効果は元素爆発1回につき1度のみ発動し、使用後15秒間も対象。
// 淡き讃歌
// 周囲のパーティメンバーが開花・超開花・烈開花・Lunar-Bloomダメージを与えた時、
// 淡き讃歌を1スタック消費し、Laumaの元素熟知に基づいてダメージが増加する。
// このダメージが複数の敵に同時に命中した場合、命中した敵の数に応じて複数のスタックが消費される。
// 各淡き讃歌スタックの持続時間は独立してカウントされる。
// Laumaが2凸以上の場合、
// 淡き讃歌の効果が強化される：周囲のパーティメンバー全員の開花・超開花・烈開花ダメージがLaumaの元素熟知の500%分さらに増加し、
// Lunar-BloomダメージがLaumaの元素熟知の400%分さらに増加する。
func (c *char) Burst(p map[string]int) (action.Info, error) {
	// 初期の淡き讃歌18スタック
	initialStacks := 18

	// 月の歌の変換を確認
	bonusStacks := 0
	if c.moonSong > 0 {
		bonusStacks = c.moonSong * 6
		c.moonSong = 0
	}

	totalStacks := initialStacks + bonusStacks
	c.paleHymn += totalStacks
	c.QueueCharTask(func() {
		c.paleHymn -= min(totalStacks, c.paleHymn) // 15秒後にスタックを削除
	}, 15*60)

	// 淡き讃歌の15秒ウィンドウの監視を設定
	c.AddStatus("pale-hymn-window", 15*60, true)

	c.SetCD(action.ActionBurst, 15*60)
	c.ConsumeEnergy(7)
	c.c1() // 1凸の元素爆発使用時効果
	c.c2() // 2凸効果の確認

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}

// 淡き讃歌の元素反応ダメージボーナスを設定
func (c *char) setupPaleHymnEffects() {

	// 開花反応をサブスクライブ
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)
		switch ae.Info.AttackTag {
		case attacks.AttackTagBloom, attacks.AttackTagHyperbloom, attacks.AttackTagBurgeon, attacks.AttackTagBountifulCore:
			return c.paleHymnReactionBonus(ae)
		case attacks.AttackTagLBDamage:
			return c.paleHymnLunarBloomBonus(ae)
		}
		return false
	}, "lauma-pale-hymn-lunar-bloom")
}

// 淡き讃歌の開花/超開花/烈開花反応ボーナス
func (c *char) paleHymnReactionBonus(ae *combat.AttackEvent) bool {
	if !c.StatusIsActive("pale-hymn-window") {
		return false
	}
	if c.paleHymn <= 0 {
		return false
	}

	targets := len(c.Core.Combat.EnemiesWithinArea(ae.Pattern, nil))
	// 敵1体ヒットにつき1スタック消費（ターゲット数を計算する必要あり）
	enemiesHit := targets
	stacksConsumed := min(c.paleHymn, enemiesHit)
	c.paleHymn -= stacksConsumed

	// 元素熟知に基づくダメージボーナスを追加
	em := c.Stat(attributes.EM)
	bonusDamage := burstBuffBloom[c.TalentLvlBurst()] * em

	// 2凸追加ボーナス
	if c.Base.Cons >= 2 {
		bonusDamage += 5.0 * em // 元素熟知の500%
	}

	ae.Info.FlatDmg += bonusDamage

	return false
}

// 淡き讃歌のLunar-Bloomダメージボーナス
func (c *char) paleHymnLunarBloomBonus(ae *combat.AttackEvent) bool {
	if !c.StatusIsActive("pale-hymn-window") {
		return false
	}
	if c.paleHymn <= 0 {
		return false
	}

	// 敵1体ヒットにつき1スタック消費
	enemiesHit := 1 // デフォルト値1
	stacksConsumed := min(c.paleHymn, enemiesHit)
	c.paleHymn -= stacksConsumed

	// 元素熟知に基づくダメージボーナスを追加
	em := c.Stat(attributes.EM)
	bonusDamage := burstBuffLBloom[c.TalentLvlBurst()] * em

	// 2凸追加ボーナス
	if c.Base.Cons >= 2 {
		bonusDamage += 4.0 * em // 元素熟知の400%
	}

	// Lunar-Bloomダメージのボーナスを保存
	ae.Info.FlatDmg += bonusDamage
	return false
}
