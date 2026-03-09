package xiangling

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var skillFrames []int

const (
	infuseWindow     = 30
	infuseDurability = 20
	particleICDKey   = "xiangling-particle-icd"
)

func init() {
	skillFrames = frames.InitAbilSlice(39)
	skillFrames[action.ActionDash] = 14
	skillFrames[action.ActionJump] = 14
	skillFrames[action.ActionSwap] = 38
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Guoba",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Pyro,
		Durability: 25,
		Mult:       guobaTick[c.TalentLvlSkill()],
	}

	// グゥオパァー消滅から固有天賦4の唐辛子取得までのフレーム遅延
	a4Delay, ok := p["a4_delay"]
	if !ok {
		a4Delay = -1
	}
	if a4Delay > 10*60 {
		a4Delay = 10 * 60
	}

	// グゥオパァーはCDフレームにスポーン
	// 7.3秒間持続、100フレームごとに発射
	c.Core.Tasks.Add(func() {
		guoba := c.newGuoba(ai)
		c.AddStatus("xianglingguoba", guoba.Duration, false)
		c.Core.Combat.AddGadget(guoba)
		// グゥオパァー消滅に対する固有天賦4をキューに追加
		if a4Delay < 0 {
			return
		}
		c.a4(guoba.Duration + a4Delay)
	}, 13)

	c.SetCDWithDelay(action.ActionSkill, 12*60, 13)

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
	c.AddStatus(particleICDKey, 1*60, false)
	c.Core.QueueParticle(c.Base.Key.String(), 1, attributes.Pyro, c.ParticleDelay)
}
