package lanyan

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/player/shield"
	"github.com/genshinsim/gcsim/pkg/core/targets"
)

var (
	skillFrames       []int
	feathermoonFrames []int
	SkillHitmarks     = []int{30, 41, 63}
)

const (
	ShieldDelay    = 14
	SkillCDDelay   = 78
	AuracheckDelay = 10
	skillWindow    = 77
	SkillKey       = "lanyan-skill-active"
	ParticleIcdKey = "lanyan-particle-icd"
)

func init() {
	skillFrames = frames.InitAbilSlice(77) // E > Anything Else
	skillFrames[action.ActionAttack] = 25  // E > FMR(E)
	skillFrames[action.ActionSkill] = 25   // E > FMR(NA)

	feathermoonFrames = frames.InitAbilSlice(61)
	feathermoonFrames[action.ActionAttack] = 52
	feathermoonFrames[action.ActionDash] = 58
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(SkillKey) {
		return c.feathermoonRing()
	}

	hold := p["hold"]
	heldtime := 0
	if hold >= 1 {
		heldtime = hold
	}

	c.DeleteStatus(ParticleIcdKey)

	c.AddStatus(SkillKey, skillWindow+heldtime, true)

	//Aura Check
	c.shieldele = attributes.Anemo
	c.QueueCharTask(func() {
		c.shieldele = c.absorbEle()
	}, AuracheckDelay+heldtime)

	//Shield
	c.QueueCharTask(func() {
		exist := c.Core.Player.Shields.Get(shield.LanyanSkill)
		if exist == nil {
			shield := shieldperct[c.TalentLvlSkill()]*c.TotalAtk() + shieldconst[c.TalentLvlSkill()]
			c.Core.Player.Shields.Add(c.newShield(shield, 12.5*60))
		} else {
			shd, _ := exist.(*shd)
			shd.Expires = c.Core.F + 12.5*60
		}
	}, ShieldDelay+heldtime)

	c.QueueCharTask(func() {
		if c.StatusIsActive(SkillKey) {
			c.SetCDWithDelay(action.ActionSkill, 16*60, 1)
		}
	}, SkillCDDelay+heldtime)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction] + heldtime,
		CanQueueAfter:   skillFrames[action.ActionAttack] + heldtime, // earliest cancel is before skillHitmark
		State:           action.SkillState,
	}, nil
}

func (c *char) particleCB(a combat.AttackCB) {
	if c.StatusIsActive(ParticleIcdKey) {
		return
	}
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	c.Core.QueueParticle(c.Base.Key.String(), 3, attributes.Anemo, c.ParticleDelay)
	c.AddStatus(ParticleIcdKey, -1, false)
}

func (c *char) absorbEle() attributes.Element {
	if c.Base.Ascension < 1 {
		return attributes.Anemo
	}
	AbsorbCheckLocation := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 1.5)
	absorbCheck := c.Core.Combat.AbsorbCheck(c.Index, AbsorbCheckLocation, attributes.Pyro, attributes.Hydro, attributes.Electro, attributes.Cryo)
	if absorbCheck == attributes.NoElement {
		return attributes.Anemo
	}
	c.IsAbsorbed = true
	return absorbCheck
}

func (c *char) feathermoonRing() (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		AttackTag:  attacks.AttackTagElementalArt,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Durability: 25,
	}
	// C1 Ring Increase
	ringnum := 1
	if c.Base.Cons >= 1 && c.IsAbsorbed {
		ringnum = 2
	}
	// A4 Flat DMG
	if c.Base.Ascension >= 4 {
		ai.FlatDmg = 3.09 * c.Stat(attributes.EM)
	}
	//Feathermoon Ring DMG
	for j := 0; j < ringnum; j++ {
		for i := 0; i < 3; i++ {
			ai.Abil = "Swallow-Wisp Pinion Dance: Feathermoon Ring DMG"
			ai.ICDTag = attacks.ICDTagElementalArt
			ai.Element = attributes.Anemo
			ai.Mult = skill[c.TalentLvlSkill()]
			c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 3), 0, SkillHitmarks[i], c.particleCB)

			//A1 Additional DMG
			if c.Base.Ascension >= 1 && c.IsAbsorbed {
				ai.Abil = "Swallow-Wisp Pinion Dance: Feathermoon Ring DMG (A4 / Absorbed)"
				ai.ICDTag = attacks.ICDTagExtraAttack
				ai.Element = c.shieldele
				ai.Mult = skill[c.TalentLvlSkill()] * 0.5
				c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 3), 0, SkillHitmarks[i], c.particleCB)
			}
		}
	}
	c.SetCDWithDelay(action.ActionSkill, 16*60, 1)
	c.DeleteStatus(SkillKey)
	return action.Info{
		Frames:          frames.NewAbilFunc(feathermoonFrames),
		AnimationLength: feathermoonFrames[action.InvalidAction],
		CanQueueAfter:   feathermoonFrames[action.ActionAttack], // earliest cancel is before skillHitmark
		State:           action.SkillState,
	}, nil
}
