package venti

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var skillPressFrames []int
var skillHoldFrames []int

func init() {
	// skill (press) -> x
	skillPressFrames = frames.InitAbilSlice(98)
	skillPressFrames[action.ActionAttack] = 22
	skillPressFrames[action.ActionAim] = 22   // 推定値
	skillPressFrames[action.ActionSkill] = 22 // 元素爆発フレームを使用
	skillPressFrames[action.ActionBurst] = 22
	skillPressFrames[action.ActionDash] = 22
	skillPressFrames[action.ActionJump] = 22

	// skill (hold) -> x
	skillHoldFrames = frames.InitAbilSlice(289)
	skillHoldFrames[action.ActionHighPlunge] = 116
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex:   c.Index,
		Abil:         "Skyward Sonnett",
		AttackTag:    attacks.AttackTagElementalArt,
		ICDTag:       attacks.ICDTagNone,
		ICDGroup:     attacks.ICDGroupDefault,
		StrikeType:   attacks.StrikeTypePierce,
		Element:      attributes.Anemo,
		Durability:   50,
		Mult:         skillPress[c.TalentLvlSkill()],
		HitWeakPoint: true,
	}

	act := action.Info{
		Frames:          frames.NewAbilFunc(skillPressFrames),
		AnimationLength: skillPressFrames[action.InvalidAction],
		CanQueueAfter:   skillPressFrames[action.ActionDash], // 最速キャンセル
		State:           action.SkillState,
	}

	cd := 360
	cdstart := 21
	hitmark := 51
	radius := 3.0
	trg := c.Core.Combat.PrimaryTarget()
	var count float64 = 3
	if p["hold"] != 0 {
		cd = 900
		cdstart = 34
		hitmark = 74
		radius = 6
		trg = c.Core.Combat.Player()
		count = 4
		ai.Mult = skillHold[c.TalentLvlSkill()]

		act = action.Info{
			Frames:          frames.NewAbilFunc(skillHoldFrames),
			AnimationLength: skillHoldFrames[action.InvalidAction],
			CanQueueAfter:   skillHoldFrames[action.ActionHighPlunge], // 最速キャンセル
			State:           action.SkillState,
		}
	} else if c.Base.Cons >= 2 && c.isHexerei {
		// 2凸 Hexerei追加：単押し「高天の歌」がオリジナルの300%のダメージを与える
		ai.Mult *= 3.0
	}

	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(trg, nil, radius), 0, hitmark, c.c2, c.makeParticleCB(count))

	c.SetCDWithDelay(action.ActionSkill, cd, cdstart)

	// 4凸（Hexerei）：Venti + チームがスキル使用後10秒間風元素ダメージ+25%
	if c.Base.Cons >= 4 {
		c.c4New()
	}

	return act, nil
}

func (c *char) makeParticleCB(count float64) combat.AttackCBFunc {
	done := false
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if done {
			return
		}
		done = true
		c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Anemo, c.ParticleDelay)
	}
}
