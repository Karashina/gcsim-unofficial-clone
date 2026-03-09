package durin

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
)

// フレームデータ: 各元素爆発のアニメーションロックとキャンセル可能タイミング
var (
	burstFrames         []int // 純化の真理のフレームデータ
	burstFramesDarkness []int // 暗闇の真理のフレームデータ
)

// タイミング定数 - 純化の真理
const (
	burstPurityHitmark1       = 98  // 1段目ヒットマーク
	burstPurityHitmark2       = 122 // 2段目ヒットマーク
	burstPurityHitmark3       = 148 // 3段目ヒットマーク
	burstPurityFirstDragonHit = 157 // 白焔の龍の初回ヒット
	dragonWhiteFlameInterval  = 59  // 白焔の龍の攻撃間隔
)

// タイミング定数 - 暗闇の真理
const (
	burstDarknessHitmark1       = 87  // 1段目ヒットマーク
	burstDarknessHitmark2       = 128 // 2段目ヒットマーク
	burstDarknessHitmark3       = 154 // 3段目ヒットマーク
	burstDarknessFirstDragonHit = 175 // 暗蝕の龍の初回ヒット
	dragonDarkDecayInterval     = 74  // 暗蝕の龍の攻撃間隔
)

// 元素爆発共通定数
const (
	burstCD        = 18 * 60 // 元素爆発クールダウン: 18秒
	dragonDuration = 20 * 60 // 龍の持続時間: 20秒
)

// 龍ステートキー: シミュレータ内部で使用するステータス識別子
const (
	dragonWhiteFlameKey = "durin-dragon-white-flame" // 白焔の龍ステート
	dragonDarkDecayKey  = "durin-dragon-dark-decay"  // 暗蝕の龍ステート
)

func init() {
	// 純化の真理のフレーム
	burstFrames = frames.InitAbilSlice(122)

	// 暗闇の真理のフレーム
	burstFramesDarkness = frames.InitAbilSlice(103)
}

// Burst は元素爆発のエントリーポイント
// 変容状態に基づいて純化の真理または暗闇の真理を発動
func (c *char) Burst(p map[string]int) (action.Info, error) {
	if c.stateDenial {
		return c.burstDarkness(p) // 暗黒の否定状態 → 暗闇の真理
	}
	return c.burstPurity(p) // 純化の肯定状態 → 純化の真理
}

// makeBurstAttackInfo は元素爆発の攻撃情報を生成するヘルパー関数
func (c *char) makeBurstAttackInfo(abilName string, mult []float64) combat.AttackInfo {
	// 元素量: 1U (元素耐性: 25)
	// ICDタグ: 元素爆発 (ICDTagElementalBurst)
	// ICDグループ: 標準 (ICDGroupDefault)
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       abilName,
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		Element:    attributes.Pyro,
		Durability: 25,
		Mult:       mult[c.TalentLvlBurst()],
	}

	return ai
}

// makeBurstDarknessAttackInfo は暗闇の真理（元素爆発）の攻撃情報を生成するヘルパー関数
func (c *char) makeBurstDarknessAttackInfo(abilName string, mult []float64) combat.AttackInfo {
	// 元素量: 1U (元素耐性: 25)
	// ICDタグ: 元素爆発 (ICDTagElementalBurst)
	// ICDグループ: 標準 (ICDGroupDefault)
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       abilName,
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		Element:    attributes.Pyro,
		Durability: 25,
		Mult:       mult[c.TalentLvlBurst()],
	}

	return ai
}

// applyC6DefIgnore は6凸（命ノ星座6）のDEF無視効果を適用するヘルパー関数
// 純化の真理: 30% DEF無視（加えて白焔の龍がヒット時にDEFデバフを適用 - cons.goで実装）
// 暗闇の真理: 70% DEF無視（基本30% + 追加40%）
func (c *char) applyC6DefIgnore(attacks []*combat.AttackInfo, isDarkness bool) {
	if c.Base.Cons < 6 {
		return
	}
	defIgnore := 0.3
	if isDarkness {
		defIgnore = 0.7 // 基本30% + 追加40%
	}
	for _, ai := range attacks {
		ai.IgnoreDefPercent = defIgnore
	}
}

// applyBurstEffects は元素爆発の共通効果を適用するヘルパー関数
// A4天賦、2凸（命ノ星座2）、エネルギー消費、クールダウン設定を処理
func (c *char) applyBurstEffects() {
	// A4: 原初融合スタックを獲得（10スタック、20秒）
	c.a4OnBurst()

	// 2凸: 元素反応バフウィンドウを有効化（20秒）
	if c.Base.Cons >= 2 {
		c.AddStatus(c2BuffKey, c2BuffDuration, true)
	}

	// 元素エネルギーを消費しクールダウンを設定
	c.ConsumeEnergy(4)
	c.SetCDWithDelay(action.ActionBurst, burstCD, 2)
}

// clearExistingDragons は新しい元素爆発使用時に既存の龍をクリアする
// 龍の効果が重複するのを防止
func (c *char) clearExistingDragons() {
	// 龍フラグをクリア
	c.dragonWhiteFlame = false
	c.dragonDarkDecay = false
	c.dragonExpiry = 0
	c.dragonSrc++ // ソースIDをインクリメントして古い龍タスクを無効化

	// ステータスキーを削除（予約済み攻撃の実行を防止）
	c.DeleteStatus(dragonWhiteFlameKey)
	c.DeleteStatus(dragonDarkDecayKey)

	c.Core.Log.NewEvent("Cleared existing dragons", glog.LogCharacterEvent, c.Index).
		Write("new_dragon_src", c.dragonSrc)
}

// burstPurity は純化の真理: 光の転回を実行
// 3回の範囲炎元素ダメージ + 白焔の龍を召喚（20秒持続、59フレームごとに範囲攻撃）
func (c *char) burstPurity(p map[string]int) (action.Info, error) {
	// 新しい龍を召喚する前に既存の龍をクリア
	c.clearExistingDragons()

	// 3つの攻撃インスタンスを生成
	ai1 := c.makeBurstAttackInfo("Principle of Purity: As the Light Shifts (Hit 1)", burstPurity1)
	ai2 := c.makeBurstAttackInfo("Principle of Purity: As the Light Shifts (Hit 2)", burstPurity2)
	ai3 := c.makeBurstAttackInfo("Principle of Purity: As the Light Shifts (Hit 3)", burstPurity3)

	// 6凸: DEF無視を適用
	c.applyC6DefIgnore([]*combat.AttackInfo{&ai1, &ai2, &ai3}, false)

	// 攻撃をキューに追加
	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 2}, 5.0)
	c.Core.QueueAttack(ai1, ap, burstPurityHitmark1, burstPurityHitmark1)
	c.Core.QueueAttack(ai2, ap, burstPurityHitmark2, burstPurityHitmark2)
	c.Core.QueueAttack(ai3, ap, burstPurityHitmark3, burstPurityHitmark3)

	// 龍を召喚し効果を適用
	c.summonDragonWhiteFlame()
	c.applyBurstEffects()

	// 1凸: 他のパーティメンバーに啓示のサイクルを付与
	if c.Base.Cons >= 1 {
		c.c1OnBurstPurity()
	}

	c.Core.Log.NewEvent("Durin uses Principle of Purity: As the Light Shifts", glog.LogCharacterEvent, c.Index)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash],
		State:           action.BurstState,
	}, nil
}

// burstDarkness は暗闇の真理: 星々の燻りを実行
// 3回の範囲炎元素ダメージ + 暗蝕の龍を召喚（20秒持続、74フレームごとに単体攻撃）
func (c *char) burstDarkness(p map[string]int) (action.Info, error) {
	// 新しい龍を召喚する前に既存の龍をクリア
	c.clearExistingDragons()

	// 3つの攻撃インスタンスを生成
	ai1 := c.makeBurstDarknessAttackInfo("Principle of Darkness: As the Stars Smolder (Hit 1)", burstDarkness1)
	ai2 := c.makeBurstDarknessAttackInfo("Principle of Darkness: As the Stars Smolder (Hit 2)", burstDarkness2)
	ai3 := c.makeBurstDarknessAttackInfo("Principle of Darkness: As the Stars Smolder (Hit 3)", burstDarkness3)

	// 6凸: DEF無視を適用（暗闇は70%）
	c.applyC6DefIgnore([]*combat.AttackInfo{&ai1, &ai2, &ai3}, true)

	// 攻撃をキューに追加
	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 2}, 5.0)
	c.Core.QueueAttack(ai1, ap, burstDarknessHitmark1, burstDarknessHitmark1)
	c.Core.QueueAttack(ai2, ap, burstDarknessHitmark2, burstDarknessHitmark2)
	c.Core.QueueAttack(ai3, ap, burstDarknessHitmark3, burstDarknessHitmark3)

	// 龍を召喚し効果を適用
	c.summonDragonDarkDecay()
	c.applyBurstEffects()

	// 1凸: デュリンに啓示のサイクルを付与
	if c.Base.Cons >= 1 {
		c.c1OnBurstDarkness()
	}

	c.Core.Log.NewEvent("Durin uses Principle of Darkness: As the Stars Smolder", glog.LogCharacterEvent, c.Index)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFramesDarkness),
		AnimationLength: burstFramesDarkness[action.InvalidAction],
		CanQueueAfter:   burstFramesDarkness[action.ActionDash],
		State:           action.BurstState,
	}, nil
}

// makeDragonAttackInfo は龍の攻撃情報を生成するヘルパー関数
// 6凸: 白焔の龍は30% DEF無視でヒット時にDEFデバフを適用（コールバックで処理）
// 6凸: 暗蝕の龍は70% DEF無視（基本30% + 追加40%）
func (c *char) makeDragonAttackInfo(abilName string, mult []float64, isDarkness bool) combat.AttackInfo {
	// 元素量: 1U (元素耐性: 25)
	// ICDタグ: Durin DoT (ICDTagDurinDoT)
	// ICDグループ: 時間ベース (白焰 90f, 暗蚀 120f)
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       abilName,
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagDurinDoT,
		ICDGroup:   attacks.ICDGroupDurinDoTWhiteFlame,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Pyro,
		Durability: 25,
		Mult:       mult[c.TalentLvlBurst()],
	}

	// 暗蝕の龍は120フレームICDグループを使用
	if isDarkness {
		ai.ICDGroup = attacks.ICDGroupDurinDoTDarkDecay
	}

	// 6凸: DEF無視
	if c.Base.Cons >= 6 {
		if isDarkness {
			ai.IgnoreDefPercent = 0.7 // 基本30% + 追加40%
		} else {
			ai.IgnoreDefPercent = 0.3 // 龍はヒット時にコールバック経由でDEFデバフも適用
		}
	}

	return ai
}

// summonDragonWhiteFlame は白焔の龍を召喚
// 20秒持続、59フレームごとに範囲炎元素ダメージを与える
func (c *char) summonDragonWhiteFlame() {
	c.dragonDarkDecay = false
	c.dragonWhiteFlame = true
	c.dragonExpiry = c.Core.F + dragonDuration

	c.AddStatus(dragonWhiteFlameKey, dragonDuration, true)
	c.DeleteStatus(dragonDarkDecayKey)

	// この龍インスタンスの現在のソースIDをキャプチャ
	dragonSrc := c.dragonSrc

	c.Core.Log.NewEvent("Dragon of White Flame summoned", glog.LogCharacterEvent, c.Index).
		Write("dragon_src", dragonSrc)

	// 定期攻撃を開始
	c.Core.Tasks.Add(func() {
		c.dragonWhiteFlameAttack(0, dragonSrc)
	}, burstPurityFirstDragonHit)
}

// dragonWhiteFlameAttack は白焔の龍の定期攻撃を実行
// 持続時間中は自動的に次の攻撃をスケジュール
func (c *char) dragonWhiteFlameAttack(attackNum int, src int) {
	// この龍インスタンスがまだ有効か確認
	if src != c.dragonSrc {
		return
	}
	if !c.StatusIsActive(dragonWhiteFlameKey) {
		return
	}

	ai := c.makeDragonAttackInfo("Dragon of White Flame", dragonWhiteFlameDmg, false)

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3.0),
		0,
		0,
		c.a4DragonAttackCB,
		c.c6DragonWhiteFlameCB,
	)

	c.Core.Tasks.Add(func() {
		c.dragonWhiteFlameAttack(attackNum+1, src)
	}, dragonWhiteFlameInterval)
}

// summonDragonDarkDecay は暗蝕の龍を召喚
// 20秒持続、74フレームごとに単体炎元素ダメージを与える
func (c *char) summonDragonDarkDecay() {
	c.dragonWhiteFlame = false
	c.dragonDarkDecay = true
	c.dragonExpiry = c.Core.F + dragonDuration

	c.AddStatus(dragonDarkDecayKey, dragonDuration, true)
	c.DeleteStatus(dragonWhiteFlameKey)

	// この龍インスタンスの現在のソースIDをキャプチャ
	dragonSrc := c.dragonSrc

	c.Core.Log.NewEvent("Dragon of Dark Decay summoned", glog.LogCharacterEvent, c.Index).
		Write("dragon_src", dragonSrc)

	// 定期攻撃を開始
	c.Core.Tasks.Add(func() {
		c.dragonDarkDecayAttack(0, dragonSrc)
	}, burstDarknessFirstDragonHit)
}

// dragonDarkDecayAttack は暗蝕の龍の定期攻撃を実行
// 持続時間中は自動的に次の攻撃をスケジュール
func (c *char) dragonDarkDecayAttack(attackNum int, src int) {
	// この龍インスタンスがまだ有効か確認
	if src != c.dragonSrc {
		return
	}
	if !c.StatusIsActive(dragonDarkDecayKey) {
		return
	}

	ai := c.makeDragonAttackInfo("Dragon of Dark Decay", dragonDarkDecayDmg, true)

	c.Core.QueueAttack(
		ai,
		combat.NewSingleTargetHit(c.Core.Combat.PrimaryTarget().Key()),
		0,
		0,
		c.a4DragonAttackCB,
	)

	c.Core.Tasks.Add(func() {
		c.dragonDarkDecayAttack(attackNum+1, src)
	}, dragonDarkDecayInterval)
}
