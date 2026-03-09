package beidou

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
	skillFrames       []int
	skillHitlagStages = []float64{.09, .09, .15}
	skillRadius       = []float64{6, 7, 8}
)

const (
	skillHitmark   = 23
	particleICDKey = "beidou-particle-icd"
)

func init() {
	skillFrames = frames.InitAbilSlice(45)
	skillFrames[action.ActionAttack] = 44
	skillFrames[action.ActionDash] = 24
	skillFrames[action.ActionJump] = 24
	skillFrames[action.ActionSwap] = 44
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	// 0=基礎ダメージ, 1=1倍ボーナス, 2=最大ボーナス
	counter := p["counter"]
	if counter >= 2 {
		counter = 2
		c.a4()
	}

	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Tidecaller (E)",
		AttackTag:          attacks.AttackTagElementalArt,
		ICDTag:             attacks.ICDTagNone,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeBlunt,
		PoiseDMG:           float64(100 * (counter + 1)),
		Element:            attributes.Electro,
		Durability:         50,
		Mult:               skillbase[c.TalentLvlSkill()] + skillbonus[c.TalentLvlSkill()]*float64(counter),
		HitlagFactor:       0.01,
		HitlagHaltFrames:   skillHitlagStages[counter] * 60,
		CanBeDefenseHalted: true,
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, skillRadius[counter]),
		skillHitmark,
		skillHitmark,
		c.makeParticleCB(counter),
	)

	// シールドを追加
	c.Core.Player.Shields.Add(&shield.Tmpl{
		ActorIndex: c.Index,
		Target:     c.Index,
		Src:        c.Core.F,
		ShieldType: shield.BeidouThunderShield,
		Name:       "Beidou Skill",
		HP:         shieldPer[c.TalentLvlSkill()]*c.MaxHP() + shieldBase[c.TalentLvlSkill()],
		Ele:        attributes.Electro,
		Expires:    c.Core.F + skillHitmark, // ヒットマークまで持続
	})

	c.SetCDWithDelay(action.ActionSkill, 450, 4)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash], // 最速キャンセル
		State:           action.SkillState,
	}, nil
}

func (c *char) makeParticleCB(counter int) combat.AttackCBFunc {
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if c.StatusIsActive(particleICDKey) {
			return
		}
		c.AddStatus(particleICDKey, 0.4*60, true)

		// ヒットなし=2, 1ヒット=3, 完全カウンター=4
		c.Core.QueueParticle(c.Base.Key.String(), 2+float64(counter), attributes.Electro, c.ParticleDelay)
	}
}
