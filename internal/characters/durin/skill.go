package durin

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

// フレームデータ: 各元素スキルアクションのアニメーションロックとキャンセル可能タイミング
var (
	skillEssentialFrames []int // Essential Transmutationのフレームデータ
	skillFrames          []int // 純化の肯定のフレームデータ
	skillDenialFrames    []int // 暗黒の否定のフレームデータ
)

// タイミング定数: ヒットマーク、クールダウン、持続時間
const (
	skillHitmark        = 33      // 純化の肯定のヒットマーク（フレーム）
	skillDenialHitmark1 = 15      // 暗黒の否定1段目ヒットマーク
	skillDenialHitmark2 = 25      // 暗黒の否定2段目ヒットマーク
	skillDenialHitmark3 = 35      // 暗黒の否定3段目ヒットマーク
	skillCD             = 12 * 60 // 元素スキルクールダウン: 12秒
	stateDuration       = 30 * 60 // 変容状態の持続時間: 30秒
	energyRegenICD      = 6 * 60  // エネルギー回復の内部クールダウン: 6秒
	skillParticleCount  = 4       // 元素粒子数
)

// 状態管理キー: シミュレータ内部で使用するステータス識別子
const (
	essentialTransmutationKey = "durin-essential-transmutation" // Essential Transmutationステート
	confirmationStateKey      = "durin-confirmation-state"      // 純化の肯定ステート
	denialStateKey            = "durin-denial-state"            // 暗黒の否定ステート
	skillRecastCDKey          = "durin-skill-recast-cd"         // 2回目のスキル連続使用のCD
)

func init() {
	// Essential Transmutation: E->NA: 16, E->E: 19
	skillEssentialFrames = frames.InitAbilSlice(50)
	skillEssentialFrames[action.ActionAttack] = 16
	skillEssentialFrames[action.ActionSkill] = 19
	skillEssentialFrames[action.ActionDash] = 16

	// Confirmation of Purity: Confirmation of Purity -> NA: 64
	skillFrames = frames.InitAbilSlice(64)
	skillFrames[action.ActionAttack] = 64
	skillFrames[action.ActionBurst] = 35
	skillFrames[action.ActionDash] = 30
	skillFrames[action.ActionJump] = 30
	skillFrames[action.ActionSwap] = 40

	// Denial of Darkness: Denial of Darkness -> NA: 62
	skillDenialFrames = frames.InitAbilSlice(62)
	skillDenialFrames[action.ActionAttack] = 62
	skillDenialFrames[action.ActionBurst] = 45
	skillDenialFrames[action.ActionDash] = 40
	skillDenialFrames[action.ActionJump] = 40
	skillDenialFrames[action.ActionSwap] = 50
}

// Skill は元素スキルのエントリーポイント
// Essential Transmutation状態に基づいて純化の肯定または暗黒の否定に分岐
func (c *char) Skill(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(essentialTransmutationKey) {
		c.particleIcd = false         // 新しいスキル使用のため粒子ICDをリセット
		return c.skillConfirmation(p) // Essential Transmutation中 → 純化の肯定
	}
	return c.skillEssentialTransmutation(p) // 通常状態 → Essential Transmutationに移行
}

// skillEssentialTransmutation はEssential Transmutation状態に移行
// 次のスキル使用で純化の肯定または暗黒の否定が使用可能になる
func (c *char) skillEssentialTransmutation(p map[string]int) (action.Info, error) {
	c.AddStatus(essentialTransmutationKey, stateDuration, true)
	c.stateDenial = false
	c.DeleteStatus(denialStateKey)

	c.Core.Log.NewEvent("Durin enters Essential Transmutation", glog.LogCharacterEvent, c.Index)
	c.SetCDWithDelay(action.ActionSkill, skillCD, 0)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillEssentialFrames),
		AnimationLength: skillEssentialFrames[action.InvalidAction],
		CanQueueAfter:   skillEssentialFrames[action.ActionDash],
		State:           action.SkillState,
	}, nil
}

// skillConfirmation は純化の肯定を実行: 範囲炎元素ダメージ
// 通常攻撃経由でEssential Transmutationに入った後、再度スキルを使用するとトリガー
func (c *char) skillConfirmation(p map[string]int) (action.Info, error) {
	// 元素量: 1U (元素耐性: 25)
	// ICDタグ: なし (ICDTagNone)
	// ICDグループ: なし - ICDTagNone使用時はICDGroupは無視される
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Transmutation: Confirmation of Purity",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		Element:    attributes.Pyro,
		Durability: 25,
		Mult:       skillPurity[c.TalentLvlSkill()],
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0}, 3.0),
		skillHitmark,
		skillHitmark,
		c.skillParticleCB,
		c.skillEnergyRegenCB,
	)

	c.transitionToConfirmationState()
	c.AddStatus(skillRecastCDKey, 0, true) // 再発動を使用済みとしてマーク

	c.Core.Log.NewEvent("Durin uses Confirmation of Purity", glog.LogCharacterEvent, c.Index)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash],
		State:           action.SkillState,
	}, nil
}

// skillDenialOfDarkness は暗黒の否定を実行: 単体炎元素ダメージ3連続ヒット
// 重撃経由でEssential Transmutationに入った後、再度スキルを使用するとトリガー
func (c *char) skillDenialOfDarkness(p map[string]int) (action.Info, error) {
	hitmarks := []int{skillDenialHitmark1, skillDenialHitmark2, skillDenialHitmark3}
	mults := [][]float64{skillDenial1, skillDenial2, skillDenial3}
	ap := combat.NewSingleTargetHit(c.Core.Combat.PrimaryTarget().Key())

	for i := 0; i < 3; i++ {
		ai := c.makeSkillDenialAttackInfo(i+1, mults[i])
		c.Core.QueueAttack(ai, ap, hitmarks[i], hitmarks[i], c.skillParticleCB, c.skillEnergyRegenCB)
	}

	c.transitionToDenialState()
	c.AddStatus(skillRecastCDKey, 0, true) // 再発動を使用済みとしてマーク
	c.Core.Log.NewEvent("Durin uses Denial of Darkness", glog.LogCharacterEvent, c.Index)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillDenialFrames),
		AnimationLength: skillDenialFrames[action.InvalidAction],
		CanQueueAfter:   skillDenialFrames[action.ActionDash],
		State:           action.SkillState,
	}, nil
}

// makeSkillDenialAttackInfo は暗黒の否定の各ヒットの攻撃情報を生成
// 3ヒットそれぞれに対して呼び出される
func (c *char) makeSkillDenialAttackInfo(hitNum int, mult []float64) combat.AttackInfo {
	// 元素量: 1U (元素耐性: 25)
	// ICDタグ: 元素スキル (ICDTagElementalArt)
	// ICDグループ: 標準 (ICDGroupDefault)
	return combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Transmutation: Denial of Darkness (Hit " + string(rune('0'+hitNum)) + ")",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagElementalArt,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeSlash,
		Element:    attributes.Pyro,
		Durability: 25,
		Mult:       mult[c.TalentLvlSkill()],
	}
}

// 状態遷移ヘルパー関数

// transitionToConfirmationState は純化の肯定状態に遷移
// この状態は純化の肯定使用後30秒間持続
func (c *char) transitionToConfirmationState() {
	c.DeleteStatus(essentialTransmutationKey)
	c.stateDenial = false
	c.DeleteStatus(denialStateKey)
	c.AddStatus(confirmationStateKey, stateDuration, true)
}

// transitionToDenialState は暗黒の否定状態に遷移
// この状態は暗黒の否定使用後30秒間持続し、その後純化の肯定に戻る
func (c *char) transitionToDenialState() {
	c.DeleteStatus(essentialTransmutationKey)
	c.stateDenial = true
	c.AddStatus(denialStateKey, stateDuration, true)

	// 持続時間後に純化の肯定に戻るようスケジュール
	c.Core.Tasks.Add(func() {
		if c.StatusIsActive(denialStateKey) {
			c.stateDenial = false
			c.DeleteStatus(denialStateKey)
			c.AddStatus(confirmationStateKey, -1, true)
			c.Core.Log.NewEvent("Durin reverts to Confirmation of Purity state", glog.LogCharacterEvent, c.Index)
		}
	}, stateDuration)
}

// コールバック関数: 攻撃ヒット時の追加効果

// skillParticleCB は元素スキルヒット時に元素粒子を生成
func (c *char) skillParticleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.particleIcd {
		return
	}
	c.particleIcd = true
	c.Core.QueueParticle(c.Base.Key.String(), skillParticleCount, attributes.Pyro, c.ParticleDelay)
}

// skillEnergyRegenCB は元素スキルヒット時に元素エネルギーを回復
// 6秒の内部クールダウンあり
func (c *char) skillEnergyRegenCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	// 6秒ICDを確認
	if c.Core.F-c.lastEnergyRestoreFrame < energyRegenICD {
		return
	}

	c.lastEnergyRestoreFrame = c.Core.F
	energyRegen := skillEnergyRegen[c.TalentLvlSkill()]
	c.AddEnergy("durin-skill-energy", energyRegen)

	c.Core.Log.NewEvent("Durin restores energy from skill", glog.LogCharacterEvent, c.Index).
		Write("energy", energyRegen)
}
