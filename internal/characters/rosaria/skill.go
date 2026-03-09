package rosaria

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var skillFrames []int

const (
	skillHitmark   = 24
	particleICDKey = "rosaria-particle-icd"
)

func init() {
	skillFrames = frames.InitAbilSlice(51)
	skillFrames[action.ActionDash] = 38
	skillFrames[action.ActionJump] = 40
	skillFrames[action.ActionSwap] = 50
}

// 元素スキルのダメージキュー生成
// 元素スキルのダメージキュー生成
// ロサリアが敵の背後に現れるかどうかのオプション引数 "nobehind" を含む（固有天賦A1用）。
// デフォルトは敵の背後に現れる。"nobehind=1" でA1発動を無効化
func (c *char) Skill(p map[string]int) (action.Info, error) {
	// 2ヒットにICDなし
	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Ravaging Confession (Hit 1)",
		AttackTag:          attacks.AttackTagElementalArt,
		ICDTag:             attacks.ICDTagNone,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeSpear,
		Element:            attributes.Cryo,
		Durability:         25,
		Mult:               skill[0][c.TalentLvlSkill()],
		HitlagHaltFrames:   0.06 * 60,
		HitlagFactor:       0.01,
		CanBeDefenseHalted: false,
	}

	// 簡略化のため、A1は常に1撃目で発動すると仮定
	var a1CB combat.AttackCBFunc
	if p["nobehind"] != 1 {
		a1CB = c.makeA1CB()
	}
	c1CB := c.makeC1CB()
	c4CB := c.makeC4CB()

	c.Core.QueueAttack(
		ai,
		combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: -1}, 2, 4),
		skillHitmark,
		skillHitmark,
		a1CB,
		c1CB,
		c4CB,
	)

	// ロサリアEは動的なので、2回目のスナップショットが必要
	//TODO: スナップショットタイミングを確認
	ai = combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Ravaging Confession (Hit 2)",
		AttackTag:          attacks.AttackTagElementalArt,
		ICDTag:             attacks.ICDTagNone,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeSlash,
		Element:            attributes.Cryo,
		Durability:         25,
		Mult:               skill[1][c.TalentLvlSkill()],
		HitlagHaltFrames:   0.09 * 60,
		HitlagFactor:       0.01,
		CanBeDefenseHalted: true,
	}
	c.QueueCharTask(func() {
		// 2撃目は1撃目の14フレーム後（ヒットラグ除外時）
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0.5}, 2.8),
			0,
			0,
			c.particleCB, // 2撃目のヒット後に粒子が生成される
			c1CB,
			c4CB,
		)
	}, skillHitmark+14)

	c.SetCDWithDelay(action.ActionSkill, 360, 23)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash], // 最速キャンセル
		State:           action.SkillState,
	}, nil
}

func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, 0.6*60, true)
	c.Core.QueueParticle(c.Base.Key.String(), 3, attributes.Cryo, c.ParticleDelay)
}
