package yunjin

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/shield"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var (
	skillFrames   [][]int
	skillHitmarks = []int{13, 50, 93}
	skillCDStarts = []int{11, 48, 90}
)

const particleICDKey = "yunjin-particle-icd"

func init() {
	skillFrames = make([][]int, 3)

	// 元素スキル（単押し）
	skillFrames[0] = frames.InitAbilSlice(62) // Tap E -> N1/Q
	skillFrames[0][action.ActionDash] = 49    // Tap E -> D
	skillFrames[0][action.ActionJump] = 48    // Tap E -> J
	skillFrames[0][action.ActionSwap] = 59    // Tap E -> Swap

	// 長押しE Lv.1
	skillFrames[1] = frames.InitAbilSlice(97) // Hold E Lv. 1 -> Q
	skillFrames[1][action.ActionAttack] = 96  // Hold E Lv. 1 -> N1
	skillFrames[1][action.ActionDash] = 85    // Hold E Lv. 1 -> D
	skillFrames[1][action.ActionJump] = 85    // Hold E Lv. 1 -> J
	skillFrames[1][action.ActionSwap] = 95    // Hold E Lv. 1 -> Swap

	// 長押しE Lv.2
	skillFrames[2] = frames.InitAbilSlice(141) // Hold E Lv. 2 -> Q
	skillFrames[2][action.ActionAttack] = 140  // Hold E Lv. 2 -> N1
	skillFrames[2][action.ActionDash] = 129    // Hold E Lv. 2 -> D
	skillFrames[2][action.ActionJump] = 129    // Hold E Lv. 2 -> J
	skillFrames[2][action.ActionSwap] = 138    // Hold E Lv. 2 -> Swap
}

// 元素スキル - 北斗のスキルをモデルにしている
// 2つのパラメータを持つ：
// perfect = 1 パーフェクトカウンター実行時
// hold = 1 または 2 通常のチャージレベル1または2
func (c *char) Skill(p map[string]int) (action.Info, error) {
	// Holdパラメータはアクションフレームで最速のリリースフレームを取得するために使用
	chargeLevel := p["hold"]
	if chargeLevel > 2 {
		chargeLevel = 2
	}
	animIdx := chargeLevel
	if p["perfect"] == 1 {
		animIdx = 0
		chargeLevel = 2
	}

	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Opening Flourish Press (E)",
		AttackTag:          attacks.AttackTagElementalArt,
		ICDTag:             attacks.ICDTagNone,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeSpear,
		Element:            attributes.Geo,
		Durability:         50,
		Mult:               skillDmg[chargeLevel][c.TalentLvlSkill()],
		UseDef:             true,
		HitlagFactor:       0.01,
		CanBeDefenseHalted: true,
	}

	var count float64
	hitDelay := skillHitmarks[animIdx]
	radius := 4.0
	switch chargeLevel {
	case 0:
		ai.HitlagHaltFrames = 0.06 * 60
		count = 2
	case 1:
		// 2または3、1:1の比率
		if c.Core.Rand.Float64() < 0.5 {
			count = 2
		} else {
			count = 3
		}
		ai.Abil = "Opening Flourish Level 1 (E)"
		ai.HitlagHaltFrames = 0.09 * 60
		radius = 6
	case 2:
		count = 3
		ai.Durability = 100
		ai.Abil = "Opening Flourish Level 2 (E)"
		ai.HitlagHaltFrames = 0.12 * 60
		radius = 8
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, radius),
		hitDelay,
		hitDelay,
		c.makeParticleCB(count),
	)

	// スキル発動までシールドを追加（攻撃命中フレームとして扱う）
	c.Core.Player.Shields.Add(&shield.Tmpl{
		ActorIndex: c.Index,
		Target:     c.Index,
		Src:        c.Core.F,
		Name:       "Yun Jin Skill",
		ShieldType: shield.YunjinSkill,
		HP:         skillShieldPct[c.TalentLvlSkill()]*c.MaxHP() + skillShieldFlat[c.TalentLvlSkill()],
		Ele:        attributes.Geo,
		Expires:    c.Core.F + hitDelay,
	})

	if c.Base.Cons >= 1 {
		// 18%は整数にならない - 442.8フレーム。切り上げ
		c.SetCDWithDelay(action.ActionSkill, 443, skillCDStarts[animIdx])
	} else {
		c.SetCDWithDelay(action.ActionSkill, 9*60, skillCDStarts[animIdx])
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames[animIdx]),
		AnimationLength: skillFrames[animIdx][action.InvalidAction],
		CanQueueAfter:   skillFrames[animIdx][action.ActionJump], // 最速キャンセル
		State:           action.SkillState,
	}, nil
}

func (c *char) makeParticleCB(count float64) combat.AttackCBFunc {
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if c.StatusIsActive(particleICDKey) {
			return
		}
		c.AddStatus(particleICDKey, 0.3*60, true)
		c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Geo, c.ParticleDelay)
	}
}
