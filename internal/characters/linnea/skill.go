package linnea

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var skillFrames []int
var skillMashFrames []int

const (
	skillHitmark     = 25
	skillMashHitmark = 35
)

func init() {
	skillFrames = frames.InitAbilSlice(35)
	skillMashFrames = frames.InitAbilSlice(55)
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	// mash=1: ルミがスーパーパワーフォーム中にアルティメットパワーフォームへ移行
	if p["mash"] == 1 && c.lumiActive && c.lumiForm == lumiFormSuper {
		return c.skillMash(p)
	}

	// 通常発動: ルミを召喚してスーパーパワーフォームに入る
	c.summonLumi(lumiFormSuper)

	// C1: Field Catalogスタックを追加
	if c.Base.Cons >= 1 {
		c.c1OnSkillUse()
	}

	// 中断耐性を増加
	c.AddStatus("linnea-poise", lumiDuration, true)

	// クールダウンを設定
	c.SetCDWithDelay(action.ActionSkill, skillCD, skillHitmark)

	c.Core.Log.NewEvent("Linnea summons Lumi in Super Power Form", glog.LogCharacterEvent, c.Index).
		Write("form", "super").
		Write("duration", lumiDuration)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash],
		State:           action.SkillState,
	}, nil
}

// skillMash はルミをアルティメットパワーフォームに移行させ、ミリオントンクラッシュを発動する
func (c *char) skillMash(p map[string]int) (action.Info, error) {
	// ミリオントンクラッシュのダメージ（Lunar-Crystallize反応ダメージ）
	ai := combat.AttackInfo{
		ActorIndex:       c.Index,
		Abil:             "Lumi Million Ton Crush (Lunar-Crystallize)",
		AttackTag:        attacks.AttackTagLCrsDamage,
		ICDTag:           attacks.ICDTagNone,
		ICDGroup:         attacks.ICDGroupDefault,
		StrikeType:       attacks.StrikeTypeBlunt,
		Element:          attributes.Geo,
		Durability:       0,
		IgnoreDefPercent: 1,
	}

	// Lunar-Crystallize式による防御力スケーリング
	mult := skillMillionTonCrush[c.TalentLvlSkill()]
	em := c.Stat(attributes.EM)
	baseDmg := c.TotalDef(false) * mult
	emBonus := (6 * em) / (2000 + em)
	ai.FlatDmg = baseDmg * (1 + c.LCrsBaseReactBonus(ai)) * (1 + emBonus + c.LCrsReactBonus(ai))
	ai.FlatDmg *= (1 + c.ElevationBonus(ai))

	// C1: Field Catalogのスタック消費 (ミリオントンクラッシュ用)
	if c.Base.Cons >= 1 {
		ai.FlatDmg += c.c1MillionTonCrushBonus()
	}

	// C2: ミリオントンクラッシュの会心ダメージ増加
	snap := combat.Snapshot{CharLvl: c.Base.Level}
	snap.Stats[attributes.CR] = c.Stat(attributes.CR)
	snap.Stats[attributes.CD] = c.Stat(attributes.CD)
	if c.Base.Cons >= 2 {
		snap.Stats[attributes.CD] += 1.50
	}

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 8)

	c.QueueCharTask(func() {
		c.Core.QueueAttackWithSnap(ai, snap, ap, 0, c.particleCB)
	}, skillMashHitmark)

	// Moonsign: Ascendant Gleamの場合、Moondrift Harmonyを発動
	if c.MoonsignAscendant {
		c.QueueCharTask(func() {
			c.triggerMoondriftHarmony()
		}, skillMashHitmark+1)
	}

	// スタンダードパワーフォームに移行
	c.lumiForm = lumiFormStandard
	c.lumiComboIdx = 0
	// ティックレートをスタンダードに変更
	c.lumiTickSrc = c.Core.F
	c.startLumiTicks(c.lumiTickSrc)

	c.Core.Log.NewEvent("Lumi uses Million Ton Crush, switching to Standard Form",
		glog.LogCharacterEvent, c.Index)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillMashFrames),
		AnimationLength: skillMashFrames[action.InvalidAction],
		CanQueueAfter:   skillMashFrames[action.ActionDash],
		State:           action.SkillState,
	}, nil
}

// summonLumi はルミを召喚する
func (c *char) summonLumi(form lumiForm) {
	c.lumiActive = true
	c.lumiSrc = c.Core.F
	c.lumiForm = form
	c.lumiComboIdx = 0

	c.AddStatus(lumiKey, lumiDuration, true)

	// ルミの定期攻撃を開始
	c.lumiTickSrc = c.Core.F
	c.startLumiTicks(c.lumiTickSrc)

	// 持続時間終了時にルミを退場させる
	src := c.lumiSrc
	c.QueueCharTask(func() {
		if c.lumiSrc != src {
			return // リセットにより無効化
		}
		c.dismissLumi()
	}, lumiDuration)
}

// dismissLumi はルミを退場させる
func (c *char) dismissLumi() {
	if !c.lumiActive {
		return
	}
	c.lumiActive = false
	c.lumiSrc = -1
	c.lumiForm = lumiFormNone
	c.lumiTickSrc = -1
	c.lumiComboIdx = 0
	c.DeleteStatus(lumiKey)

	c.Core.Log.NewEvent("Lumi dismissed", glog.LogCharacterEvent, c.Index)
}

// resetLumiDuration はルミの持続時間をリセットする（形態は変更しない）
func (c *char) resetLumiDuration() {
	if !c.lumiActive {
		return
	}
	src := c.Core.F
	c.lumiSrc = src
	c.AddStatus(lumiKey, lumiDuration, true)

	// 退場タスクを再スケジュール
	c.QueueCharTask(func() {
		if c.lumiSrc != src {
			return
		}
		c.dismissLumi()
	}, lumiDuration)

	c.Core.Log.NewEvent("Lumi duration reset", glog.LogCharacterEvent, c.Index).
		Write("form", c.lumiForm)
}

// startLumiTicks はルミの定期攻撃ティックを開始する
func (c *char) startLumiTicks(src int) {
	tickRate := lumiSuperTickRate
	if c.lumiForm == lumiFormStandard {
		tickRate = lumiStandardTickRate
	}

	c.QueueCharTask(func() {
		if c.lumiTickSrc != src {
			return // 無効化
		}
		if !c.lumiActive {
			return
		}

		c.lumiAttackTick()

		// 次のティックをスケジュール
		c.startLumiTicks(src)
	}, tickRate)
}

// lumiAttackTick はルミの1回の攻撃ティックを実行する
func (c *char) lumiAttackTick() {
	switch c.lumiForm {
	case lumiFormSuper:
		c.lumiSuperFormAttack()
	case lumiFormStandard:
		c.lumiStandardFormAttack()
	default:
		// アルティメットフォームはskillMashで直接処理
		return
	}
}

// lumiSuperFormAttack はスーパーパワーフォームの攻撃を実行する
// Moondriftsがある場合: パンチx2 → ハンマーx1 のサイクル
// Moondriftsがない場合: パンチx2 のみ
func (c *char) lumiSuperFormAttack() {
	hasMoondrifts := c.MoonsignAscendant

	if hasMoondrifts && c.lumiComboIdx == 2 {
		// ヘビーオーバードライブハンマー（Lunar-Crystallize反応ダメージ）
		c.lumiHeavyOverdriveHammer()
		c.lumiComboIdx = 0

		// Moonsign: Ascendant Gleamの場合、Moondrift Harmonyを発動
		c.triggerMoondriftHarmony()
	} else {
		// パウンドパウンドパンメラー（2ヒット岩元素ダメージ）
		c.lumiPoundPoundPummeler()
		c.lumiComboIdx++
		if !hasMoondrifts {
			c.lumiComboIdx = 0 // Moondriftsなしの場合はリセット
		}
	}
}

// lumiStandardFormAttack はスタンダードパワーフォームの攻撃を実行する
func (c *char) lumiStandardFormAttack() {
	c.lumiPoundPoundPummeler()
}

// lumiPoundPoundPummeler はパウンドパウンドパンメラー攻撃を実行する（2ヒット）
func (c *char) lumiPoundPoundPummeler() {
	for i := 0; i < 2; i++ {
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Lumi Pound-Pound Pummeler",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagElementalArt,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeBlunt,
			Element:    attributes.Geo,
			Durability: 25,
			UseDef:     true,
			Mult:       skillPoundPound[c.TalentLvlSkill()],
		}

		delay := i * 10 // 2ヒットの間隔
		ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5)
		c.QueueCharTask(func() {
			c.Core.QueueAttack(ai, ap, 0, 0, c.particleCB)
		}, delay)
	}
}

// lumiHeavyOverdriveHammer はヘビーオーバードライブハンマー攻撃を実行する
// （Lunar-Crystallize反応ダメージ）
func (c *char) lumiHeavyOverdriveHammer() {
	ai := combat.AttackInfo{
		ActorIndex:       c.Index,
		Abil:             "Lumi Heavy Overdrive Hammer (Lunar-Crystallize)",
		AttackTag:        attacks.AttackTagLCrsDamage,
		ICDTag:           attacks.ICDTagNone,
		ICDGroup:         attacks.ICDGroupDefault,
		StrikeType:       attacks.StrikeTypeBlunt,
		Element:          attributes.Geo,
		Durability:       0,
		IgnoreDefPercent: 1,
	}

	// Lunar-Crystallize式による防御力スケーリング
	mult := skillHeavyOverdrive[c.TalentLvlSkill()]
	em := c.Stat(attributes.EM)
	baseDmg := c.TotalDef(false) * mult
	emBonus := (6 * em) / (2000 + em)
	ai.FlatDmg = baseDmg * (1 + c.LCrsBaseReactBonus(ai)) * (1 + emBonus + c.LCrsReactBonus(ai))
	ai.FlatDmg *= (1 + c.ElevationBonus(ai))

	snap := combat.Snapshot{CharLvl: c.Base.Level}
	snap.Stats[attributes.CR] = c.Stat(attributes.CR)
	snap.Stats[attributes.CD] = c.Stat(attributes.CD)

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 6)
	c.Core.QueueAttackWithSnap(ai, snap, ap, 0)
}

// particleCB はパーティクル生成用コールバック（ICD付き）
func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, 2*60, true)

	count := 3.0
	c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Geo, c.ParticleDelay)
}
