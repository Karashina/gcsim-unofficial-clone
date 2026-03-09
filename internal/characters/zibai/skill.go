package zibai

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var skillFrames []int
var spiritSteedFrames []int

const (
	skillHitmark        = 22
	spiritSteedHitmark1 = 31
	spiritSteedHitmark2 = 35
)

func init() {
	skillFrames = frames.InitAbilSlice(32)

	spiritSteedFrames = frames.InitAbilSlice(54)
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	// 既に月相転移モード中で十分な輝度がある場合、神馬駆けを使用
	if c.lunarPhaseShiftActive && c.phaseShiftRadiance >= spiritSteedRadianceCost {
		return c.spiritSteedStride(p)
	}

	// 月相転移モードに入る
	c.enterLunarPhaseShift()

	c.Core.Log.NewEvent("Zibai enters Lunar Phase Shift mode", glog.LogCharacterEvent, c.Index).
		Write("duration", lunarPhaseShiftDuration)

	// クールダウンを設定（18秒）
	c.SetCDWithDelay(action.ActionSkill, 18*60, skillHitmark)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash],
		State:           action.SkillState,
	}, nil
}

// enterLunarPhaseShift は月相転移モードを発動する
func (c *char) enterLunarPhaseShift() {
	c.lunarPhaseShiftActive = true
	src := c.Core.F
	c.lunarPhaseShiftSrc = src

	// 1凸: 即座に月相転移輝度100を獲得
	if c.Base.Cons >= 1 {
		c.phaseShiftRadiance = 100
		c.c1FirstStride = true // 初回ストライドにボーナス
	} else {
		c.phaseShiftRadiance = 0
	}

	c.spiritSteedUsages = 0
	c.AddStatus(skillKey, lunarPhaseShiftDuration, true)

	// メソッドから定期的な輝度獲得を開始
	c.startRadianceAccumulation(src)

	// モード終了をスケジュール（元素爆発延長が無効化できるようにソースチェック付き）
	c.QueueCharTask(func() {
		if c.lunarPhaseShiftSrc != src {
			return // 延長により無効化
		}
		c.exitLunarPhaseShift()
	}, lunarPhaseShiftDuration)
}

// exitLunarPhaseShift は月相転移モードを終了する
func (c *char) exitLunarPhaseShift() {
	if !c.lunarPhaseShiftActive {
		return
	}
	c.lunarPhaseShiftActive = false
	c.lunarPhaseShiftSrc = -1
	c.phaseShiftRadiance = 0
	c.spiritSteedUsages = 0
	c.c1FirstStride = false
	// 4凸がない場合、保存された通常攻撃カウンターをリセット
	if c.Base.Cons < 4 {
		c.savedNormalCounter = 0
	}
	c.DeleteStatus(skillKey)
	c.DeleteStatus(radianceNormalICDKey)
	c.DeleteStatus(radianceLCrsICDKey)

	c.Core.Log.NewEvent("Zibai exits Lunar Phase Shift mode", glog.LogCharacterEvent, c.Index)
}

// extendLunarPhaseShift は月相転移の持続時間を指定フレーム分延長する
func (c *char) extendLunarPhaseShift(extensionFrames int) {
	if !c.lunarPhaseShiftActive {
		return
	}
	c.ExtendStatus(skillKey, extensionFrames)

	// 終了タスクを再スケジュール: 古い終了を無効化するためにソースを更新し、新しいタスクをキューに追加
	src := c.Core.F
	c.lunarPhaseShiftSrc = src
	c.startRadianceAccumulation(src)

	remaining := c.StatusDuration(skillKey)
	c.QueueCharTask(func() {
		if c.lunarPhaseShiftSrc != src {
			return // 別の延長により無効化
		}
		c.exitLunarPhaseShift()
	}, remaining)

	c.Core.Log.NewEvent("Zibai Lunar Phase Shift extended", glog.LogCharacterEvent, c.Index).
		Write("extension_frames", extensionFrames).
		Write("remaining_frames", remaining)
}

// spiritSteedStride は神馬駆け攻撃を実行する
func (c *char) spiritSteedStride(p map[string]int) (action.Info, error) {
	// 6凸: 全輝度を消費してボーナスダメージ
	consumedRadiance := c.phaseShiftRadiance
	var c6BonusPct float64 = 0

	if c.Base.Cons >= 6 && consumedRadiance > 70 {
		// 70を超えた1ポイントにつき1.6%
		c6BonusPct = float64(consumedRadiance-70) * 0.016
		c.applyC6ElevationBuff(c6BonusPct)
	}

	// 輝度を消費（6凸は全消費、それ以外は70）
	if c.Base.Cons >= 6 {
		c.phaseShiftRadiance = 0
	} else {
		c.phaseShiftRadiance -= spiritSteedRadianceCost
	}
	c.spiritSteedUsages++
	c.DeleteStatus(radianceLCrsICDKey)

	// 1段目ダメージ
	ai1 := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Spirit Steed's Stride 1-Hit",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Geo,
		Durability: 25,
		UseDef:     true,
		Mult:       spiritSteedStride_1[c.TalentLvlSkill()],
	}

	ap1 := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5)

	c.QueueCharTask(func() {
		c.Core.QueueAttack(ai1, ap1, 0, 0, c.spiritSteedOnHitCB)
	}, spiritSteedHitmark1)

	// 2段目ヒットの倍率をボーナス込みで計算
	secondHitMult := 1.6 * spiritSteedStride_2[c.TalentLvlSkill()]

	ai2 := combat.AttackInfo{
		ActorIndex:       c.Index,
		Abil:             "Spirit Steed's Stride 2-Hit (Lunar-Crystallize / E)",
		AttackTag:        attacks.AttackTagLCrsDamage,
		ICDTag:           attacks.ICDTagNone,
		ICDGroup:         attacks.ICDGroupDefault,
		StrikeType:       attacks.StrikeTypeDefault,
		Element:          attributes.Geo,
		Durability:       0,
		IgnoreDefPercent: 1,
	}

	// 1凸: 初回神馬駆けの2段目が220%増加
	c1bonus := 0.0
	if c.c1FirstStride {
		c1bonus = 2.2
		c.c1FirstStride = false
	}

	// Lunar-Crystallize式による防御力スケーリング
	em := c.Stat(attributes.EM)
	baseDmg := c.TotalDef(false) * secondHitMult
	emBonus := (6 * em) / (2000 + em)
	ai2.FlatDmg = baseDmg * (1 + c.LCrsBaseReactBonus(ai2) + c1bonus) * (1 + emBonus + c.LCrsReactBonus(ai2))

	// 固有天賦1: Selenic Descent効果 - 2段目を防御力の60%分増加
	// 2凸: 月相がAscendant Gleamの時、固有天賦1がさらに防御力の550%増加
	// （追加固定ダメージ、asc.goで処理）
	if c.StatusIsActive(selenicDescentKey) {
		if c.Base.Cons >= 2 && c.isMoonsignAscendant() {
			ai2.FlatDmg += 5.5 * c.TotalDef(false)
		} else {
			ai2.FlatDmg += 0.6 * c.TotalDef(false)
		}
	}

	ai2.FlatDmg *= (1 + c.ElevationBonus(ai2))

	snap := combat.Snapshot{CharLvl: c.Base.Level}
	snap.Stats[attributes.CR] = c.Stat(attributes.CR)
	snap.Stats[attributes.CD] = c.Stat(attributes.CD)

	ap2 := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 8)
	c.QueueCharTask(func() {
		c.Core.QueueAttackWithSnap(ai2, snap, ap2, 0, c.spiritSteedOnHitCB)
	}, spiritSteedHitmark2)

	c.Core.Log.NewEvent("Zibai uses Spirit Steed's Stride", glog.LogCharacterEvent, c.Index).
		Write("radiance_consumed", consumedRadiance).
		Write("usages", c.spiritSteedUsages).
		Write("c6_bonus_pct", c6BonusPct)

	// 仕様: 4回使用（無凸）または最大使用回数（1凸: 5回）後に月相転移を終了
	if c.spiritSteedUsages >= c.maxSpiritSteedUsages {
		c.QueueCharTask(func() {
			c.exitLunarPhaseShift()
		}, spiritSteedHitmark2+1) // 2段目の着弾後に終了
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(spiritSteedFrames),
		AnimationLength: spiritSteedFrames[action.InvalidAction],
		CanQueueAfter:   spiritSteedFrames[action.ActionDash],
		State:           action.SkillState,
	}, nil
}

// spiritSteedOnHitCB は神馬駆けの命中時コールバック
func (c *char) spiritSteedOnHitCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	// 4凸: Scattermoon Splendor効果を獲得
	if c.Base.Cons >= 4 {
		c.c4ScattermoonUsed = true
	}
}

func (c *char) initRadianceHandlers() {
	c.Core.Events.Subscribe(event.OnLunarCrystallize, func(args ...interface{}) bool {
		if !c.lunarPhaseShiftActive {
			return false
		}
		if c.StatusIsActive(radianceLCrsICDKey) {
			return false
		}
		c.AddStatus(radianceLCrsICDKey, radianceLCrsICD, false)
		c.addPhaseShiftRadiance(radianceLCrsGain)
		return false
	}, "zibai-radiance-lcrs")
}

// startRadianceAccumulation は定期的な輝度蓄積を開始する
func (c *char) startRadianceAccumulation(src int) {
	// 時間経過で輝度を獲得
	c.QueueCharTask(func() {
		if c.lunarPhaseShiftSrc != src {
			return
		}
		if !c.lunarPhaseShiftActive {
			return
		}
		c.addPhaseShiftRadiance(radianceTickGain)
		c.startRadianceAccumulation(src)
	}, radianceTickInterval)
}

// particleCBは粒子生成を処理
func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, 2*60, true)
	c.Core.QueueParticle(c.Base.Key.String(), 1, attributes.Geo, c.ParticleDelay)
}
