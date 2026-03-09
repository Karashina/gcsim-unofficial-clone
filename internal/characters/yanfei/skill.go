package yanfei

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

const (
	skillHitmark   = 32
	particleICDKey = "yanfei-particle-icd"
)

func init() {
	skillFrames = frames.InitAbilSlice(46) // E -> N1
	skillFrames[action.ActionCharge] = 35  // E -> CA
	skillFrames[action.ActionBurst] = 43   // E -> Q
	skillFrames[action.ActionDash] = 29    // E -> D
	skillFrames[action.ActionJump] = 34    // E -> J
	skillFrames[action.ActionSwap] = 44    // E -> Swap
}

// 煙绯の元素スキル - 他のキットとの相互作用が少なく単純
// 炎を召喚して範囲炎元素ダメージを与える。ヒットした敵は煙绯に最大数の朱印を付与する。
func (c *char) Skill(p map[string]int) (action.Info, error) {
	done := false
	addSeal := func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if done {
			return
		}
		// ヒット時に朱印を最大数付与
		if c.sealCount < c.maxTags {
			c.sealCount = c.maxTags
		}
		c.AddStatus(sealBuffKey, 600, true)
		c.Core.Log.NewEvent("yanfei gained max seals", glog.LogCharacterEvent, c.Index).
			Write("current_seals", c.sealCount)
		done = true
	}

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Signed Edict",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		PoiseDMG:   120,
		Element:    attributes.Pyro,
		Durability: 25,
		Mult:       skill[c.TalentLvlSkill()],
	}
	// TODO: スナップショットのタイミングが不明
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHit(
			c.Core.Combat.Player(),
			c.Core.Combat.PrimaryTarget(),
			nil,
			3.5,
		),
		0,
		skillHitmark,
		c.particleCB,
		addSeal,
	)

	c.SetCDWithDelay(action.ActionSkill, 540, 28)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash], // 最速キャンセルはスキルヒットマークより前
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
	c.AddStatus(particleICDKey, 0.2*60, true)
	c.Core.QueueParticle(c.Base.Key.String(), 3, attributes.Pyro, c.ParticleDelay)
}
