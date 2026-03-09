package qiqi

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
)

var skillFrames []int

const (
	skillHitmark = 32
	skillBuffKey = "qiqi-e"
)

func init() {
	skillFrames = frames.InitAbilSlice(57) // E -> N1/Swap
	skillFrames[action.ActionBurst] = 58   // E -> Q
	skillFrames[action.ActionDash] = 6     // E -> D
	skillFrames[action.ActionJump] = 5     // E -> J
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	// 終了時刻の問題を避けるため+1
	// 七七の元素スキルは初撃後は設置物なので、ヒットラグによる延長はされない
	c.AddStatus(skillBuffKey, 15*60+1, false)
	c.skillLastUsed = c.Core.F
	src := c.Core.F

	// 初撃ダメージ
	// 回復とダメージの両方がスナップショット
	c.Core.Tasks.Add(func() {
		ai := combat.AttackInfo{
			ActorIndex:         c.Index,
			Abil:               "Herald of Frost: Initial Damage",
			AttackTag:          attacks.AttackTagElementalArt,
			ICDTag:             attacks.ICDTagElementalArt,
			ICDGroup:           attacks.ICDGroupDefault,
			StrikeType:         attacks.StrikeTypeDefault,
			Element:            attributes.Cryo,
			Durability:         25,
			Mult:               skillInitialDmg[c.TalentLvlSkill()],
			HitlagFactor:       0.05,
			HitlagHaltFrames:   0.05 * 60,
			CanBeDefenseHalted: true,
		}
		snap := c.Snapshot(&ai)

		// 発動時に1回の回復が即座に発生
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index,
			Target:  c.Core.Player.Active(),
			Message: "Herald of Frost (Tick)",
			Src:     c.healSnapshot(&snap, skillHealContPer, skillHealContFlat, c.TalentLvlSkill()),
			Bonus:   snap.Stats[attributes.Heal],
		})

		// 回復とダメージのインスタンスはスナップショット
		// 別々にクローンされたスナップショットが各関数に渡され、相互干渉を防止

		// 継続回復インスタンスをキューに追加
		// 回復Tickの正確なフレームデータはない。ここでは大まかに推定
		// 回復Tickはスキル中に追加3回発生 - Tick間隔は約4.5秒と仮定
		// 秒単位（0 = スキル発動）: 1, 5.5, 10, 14.5
		c.skillHealSnapshot = snap
		c.Core.Tasks.Add(c.skillHealTickTask(src), 4.5*60)

		// ダメージスワイプインスタンスをキューに追加。
		// ダメージTickの正確なフレームデータはない。ここでは大まかに推定
		// スキル期間中に9回発生
		// 初回は発動直後、その後残りの持続時間にわたっ8回追加発動
		// 各発動は約2.25秒間隔の「ペア」で2回ずつ発生
		// ペア内の各スワイプの間隔は約1秒
		// 正確なフレームデータがなく、スキル持続時間はヒットラグの影響を受ける
		// ダメージ発動（秒単位 0 = スキル発動）: 1.5, 3.75, 4.75, 7, 8, 10.25, 11.25, 13.5, 14.5

		aiTick := ai
		aiTick.Abil = "Herald of Frost: Skill Damage"
		aiTick.Mult = skillDmgCont[c.TalentLvlSkill()]
		aiTick.IsDeployable = true // ティックはヒットラグを適用するが設置物なので七七に影響しない

		snapTick := c.Snapshot(&aiTick)
		tickAE := &combat.AttackEvent{
			Info:        aiTick,
			Snapshot:    snapTick,
			SourceFrame: c.Core.F,
		}

		// 遮光器ヒットラグを追加遅延として仮定（元素スキルTick 1は初撃ヒットラグで遅延）
		// 他のソースからのヒットラグはカウントしないため、キャラキューは使用不可
		c.Core.Tasks.Add(c.skillDmgTickTask(src, tickAE, 60), 57+7)

		// ステータスが正しく処理されるよう、ダメージ適用は上記の後に実行する必要がある
		c.Core.QueueAttackWithSnap(ai, snap, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 2.5), 0)
	}, skillHitmark)

	c.SetCDWithDelay(action.ActionSkill, 1800, 3) // 30s * 60

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionJump], // 最速キャンセルはスキルヒットマークより前
		State:           action.SkillState,
	}, nil
}

// 元素スキルのダメージスワイプインスタンスを処理
// 1命ノ星座も処理:
// 寒病鬼差が寿命の箓が付与された敵に命中すると、七七は元素エネルギーを2回復する。
func (c *char) skillDmgTickTask(src int, ae *combat.AttackEvent, lastTickDuration int) func() {
	return func() {
		if !c.StatusIsActive(skillBuffKey) {
			return
		}

		// TODO: 祭礼の剣との相互作用が不明...一度に1つのインスタンスのみとして扱う
		if c.skillLastUsed > src {
			return
		}

		// 初期スナップショットを複製
		tick := *ae // ポインタを参照解除
		// スキルがプレイヤーに追従するため、AttackEvent生成時にスナップショットすべきではない
		tick.Pattern = combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 2.5)

		if c.Base.Cons >= 1 {
			tick.Callbacks = append(tick.Callbacks, c.c1)
		}

		c.Core.QueueAttackEvent(&tick, 0)

		nextTick := 60
		if lastTickDuration == 60 {
			nextTick = 135
		}
		c.Core.Tasks.Add(c.skillDmgTickTask(src, ae, nextTick), nextTick)
	}
}

// 元素スキルの自動回復Tickを処理
func (c *char) skillHealTickTask(src int) func() {
	return func() {
		if !c.StatusIsActive(skillBuffKey) {
			return
		}

		// TODO: 祭礼の剣との相互作用が不明...一度に1つのインスタンスのみとして扱う
		if c.skillLastUsed > src {
			return
		}

		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index,
			Target:  c.Core.Player.Active(),
			Message: "Herald of Frost (Tick)",
			Src:     c.healSnapshot(&c.skillHealSnapshot, skillHealContPer, skillHealContFlat, c.TalentLvlSkill()),
			Bonus:   c.skillHealSnapshot.Stats[attributes.Heal],
		})

		// 次のインスタンスをキューに追加
		c.Core.Tasks.Add(c.skillHealTickTask(src), 4.5*60)
	}
}
