package varka

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var (
	skillFrames []int
	fwaFrames   []int
	fwaHitmarks = []int{31, 52}
)

const (
	skillHitmark         = 41
	skillConversionStart = 50
)

func init() {
	skillFrames = frames.InitAbilSlice(65)

	fwaFrames = frames.InitAbilSlice(65)
	fwaFrames[action.ActionDash] = fwaHitmarks[1]
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	if c.sturmActive {
		// C6: FWA後、スキルタップで追加Azure Devourを発動（FWAではない）
		if c.Base.Cons >= 6 && c.StatusIsActive(c6FWAWindowKey) {
			info, err := c.azureDevour(p)
			if err == nil {
				info.State = action.SkillState
			}
			return info, err
		}
		return c.fourWindsAscension(p)
	}
	return c.windBoundExecution(p)
}

// windBoundExecution はSturm und Drangに入る初期スキル発動
func (c *char) windBoundExecution(p map[string]int) (action.Info, error) {
	// 風元素ダメージを与える（スキルダメージ）
	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Windbound Execution",
		AttackTag:          attacks.AttackTagElementalArt,
		ICDTag:             attacks.ICDTagNone,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeBlunt,
		PoiseDMG:           150.0,
		Element:            attributes.Anemo,
		Durability:         25,
		Mult:               skillDmg[c.TalentLvlSkill()],
		HitlagFactor:       0.01,
		HitlagHaltFrames:   0.1 * 60,
		CanBeDefenseHalted: true,
	}

	c.Core.QueueAttack(
		ai,
		combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, 4, 5),
		skillHitmark, skillHitmark,
		c.skillParticleCB,
	)

	// Sturm und Drangモードに入る
	c.enterSturmUndDrang()

	// holdパラメータに基づいてCDを設定
	hold, ok := p["hold"]
	if ok && hold > 0 {
		c.SetCD(action.ActionSkill, 8*60) // Hold CD = 8s
	} else {
		c.SetCD(action.ActionSkill, 16*60) // Tap CD = 16s
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillHitmark,
		State:           action.SkillState,
	}, nil
}

// enterSturmUndDrang はS&Dモードを有効化する
func (c *char) enterSturmUndDrang() {
	c.sturmActive = true
	c.sturmSrc = c.Core.F
	c.cdReductionCount = 0

	// FWAチャージ: CD中から開始ﾈ0チャージﾉ、CD = チャージあたり11秒
	c.fwaCharges = 0
	c.fwaCDEndFrame = c.Core.F + 11*60

	// C1: S&D突入時にFWAチャージを1つ即座に付与
	if c.Base.Cons >= 1 {
		c.fwaCharges = 1
	}

	// S&D持続時間: 12秒
	dur := 12 * 60
	c.AddStatus(sturmUndDrangKey, dur, true)

	// C1: Lyrical Libation - 最初のFWA/Azure Devourが200%ダメージ
	if c.Base.Cons >= 1 {
		c.AddStatus(c1LyricalKey, dur, true)
	}

	// S&D有効期限ﾈ12秒 = 720フレームﾉに終了をスケジュール。
	// Core.Tasks.Addはc.Core.F（絶対フレームカウント）を使用; QueueCharTaskを
	// 使うとヒットラグがTimePassedを一時停止するためS&Dが長くなりすぎる。
	src := c.sturmSrc
	c.Core.Tasks.Add(func() {
		if c.sturmSrc != src {
			return
		}
		c.exitSturmUndDrang()
	}, dur)
}

// exitSturmUndDrang はS&Dモードを無効化する
func (c *char) exitSturmUndDrang() {
	c.sturmActive = false
	c.fwaCharges = 0
	c.DeleteStatus(sturmUndDrangKey)
	c.DeleteStatus(c1LyricalKey)
}

// fourWindsAscension はS&Dモード中の特殊スキルを処理する
func (c *char) fourWindsAscension(p map[string]int) (action.Info, error) {
	lvl := c.TalentLvlSkill()

	// C6: FWAウィンドウからの発動か確認（チャージ消費なし）
	consumeCharge := true
	if c.Base.Cons >= 6 {
		if c.StatusIsActive(c6AzureWindowKey) {
			// Azure Devour後、スキルタップでチャージ消費なしの追加FWAを発動
			consumeCharge = false
			c.DeleteStatus(c6AzureWindowKey)
		}
		// c6FWAWindowKeyのケースはSkill()がazureDevourにルーティングすることで処理
	}

	if consumeCharge {
		c.fwaCharges--
	}

	// C1: Lyrical Libation効果
	c1Mult := 1.0
	if c.Base.Cons >= 1 && c.StatusIsActive(c1LyricalKey) {
		c1Mult = 2.0
		c.DeleteStatus(c1LyricalKey)
	}

	// FWA: 2ヒット
	// 1打目: 他元素（ICDなし）, 2打目: 風（ICDなし）
	otherEle := c.otherElement
	if !c.hasOtherEle {
		otherEle = attributes.Anemo
	}

	// ヒット1: 他元素
	mult1 := fwaOther[lvl]
	if c.Base.Ascension >= 1 && c.a1MultFactor != 1.0 {
		mult1 *= c.a1MultFactor
	}
	mult1 *= c1Mult

	ai1 := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Four Winds' Ascension (Other)",
		AttackTag:          attacks.AttackTagElementalArt,
		ICDTag:             attacks.ICDTagNone,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeBlunt,
		PoiseDMG:           150.0,
		Element:            otherEle,
		Durability:         25,
		Mult:               mult1,
		HitlagFactor:       0.01,
		HitlagHaltFrames:   0.1 * 60,
		CanBeDefenseHalted: true,
	}
	c.Core.QueueAttack(
		ai1,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, 4),
		fwaHitmarks[0], fwaHitmarks[0],
	)

	// ヒット2: 風元素
	mult2 := fwaAnemo[lvl]
	if c.Base.Ascension >= 1 && c.a1MultFactor != 1.0 {
		mult2 *= c.a1MultFactor
	}
	mult2 *= c1Mult

	ai2 := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Four Winds' Ascension (Anemo)",
		AttackTag:          attacks.AttackTagElementalArt,
		ICDTag:             attacks.ICDTagNone,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeBlunt,
		PoiseDMG:           100.0,
		Element:            attributes.Anemo,
		Durability:         25,
		Mult:               mult2,
		HitlagFactor:       0.01,
		HitlagHaltFrames:   0.08 * 60,
		CanBeDefenseHalted: true,
	}
	c.Core.QueueAttack(
		ai2,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, 4),
		fwaHitmarks[1], fwaHitmarks[1],
	)

	// C2: ATKの800%に等しい追加風元素攻撃
	if c.Base.Cons >= 2 {
		c.c2Strike(fwaHitmarks[1] + 4)
	}

	// C6: FWA後、追加Azure Devour用のウィンドウを開く
	// 通常のFWAの場合のみウィンドウを設定（C6チェーントリガーではない）
	if c.Base.Cons >= 6 && consumeCharge {
		c.AddStatus(c6FWAWindowKey, 60, true) // 約1秒のウィンドウ
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(fwaFrames),
		AnimationLength: fwaFrames[action.InvalidAction],
		CanQueueAfter:   fwaHitmarks[0],
		State:           action.SkillState,
	}, nil
}

// skillParticleCB はスキルが敵に命中した時に粒子を生成する
func (c *char) skillParticleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, 0.5*60, true)

	// 6個の元素粒子を生成
	c.Core.QueueParticle(c.Base.Key.String(), 6, attributes.Anemo, c.ParticleDelay)
}

// c2Strike はC2の追加風元素攻撃を実行する
func (c *char) c2Strike(delay int) {
	atk := c.TotalAtk()
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "C2: Dawn's Flight Strike",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		PoiseDMG:   100.0,
		Element:    attributes.Anemo,
		Durability: 25,
		FlatDmg:    atk * 8.0, // 800% ATK
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, 5),
		delay, delay,
	)
}

func (c *char) c2Init() {
	// C2はFWAとAzure Devour関数内で直接処理
}
