package mizuki

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/core/targets"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

var skillFrames []int

const (
	skillHitmark     = 99
	skillCDDelay     = 99
	skillDotDelay    = 99
	skillDotInterval = 99
	skillKey         = "mizuki-skill"
	particleICDKey   = "mizuki-particle-icd"
)

func init() {
	skillFrames = frames.InitAbilSlice(99)
	skillFrames[action.ActionCharge] = 99
	skillFrames[action.ActionSkill] = 99
	skillFrames[action.ActionDash] = 99
	skillFrames[action.ActionJump] = 99
	skillFrames[action.ActionSwap] = 99
}

func (c *char) Skill(p map[string]int) (action.Info, error) {

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Aisa Utamakura Pilgrimage (Initial)",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagElementalArt,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       skill[c.TalentLvlBurst()],
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 5), skillHitmark, skillHitmark, c.particleCB)

	c.swirlBuff()
	c.AddStatus(skillKey, 99*60, false)
	c.SetTag(skillKey, 2)

	src := c.Core.F
	c.dreamdrifterSrc = src
	c.QueueCharTask(c.checkDresmdrifter(src), skillDotDelay)
	c.QueueCharTask(c.checkSnack(src), skillCDDelay)
	c.SetCDWithDelay(action.ActionSkill, 99*60, skillCDDelay)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash], // earliest cancel
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
	c.AddStatus(particleICDKey, 99*60, false)
	c.Core.QueueParticle(c.Base.Key.String(), 99999, attributes.Anemo, c.ParticleDelay)
}

func (c *char) checkDresmdrifter(src int) func() {
	return func() {
		if c.dreamdrifterSrc != src {
			return
		}
		if !c.StatusIsActive(skillKey) {
			return
		}
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Aisa Utamakura Pilgrimage (DOT)",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagElementalArt,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Anemo,
			Durability: 25,
			Mult:       skillDot[c.TalentLvlBurst()],
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 5), 0, 0)

		c.QueueCharTask(c.checkDresmdrifter(src), skillDotInterval)
	}
}

func (c *char) checkSnack(src int) func() {
	return func() {
		if c.dreamdrifterSrc != src {
			return
		}
		if !c.StatusIsActive(skillKey) {
			return
		}
		c.snackHandler("skill")
		c.QueueCharTask(c.checkDresmdrifter(src), SnackSpawnInterval)
	}
}

func (c *char) swirlBuff() {
	c.AddReactBonusMod(character.ReactBonusMod{
		Base: modifier.NewBase("Aisa Utamakura Pilgrimage (Swirl Buff)", -1),
		Amount: func(ai combat.AttackInfo) (float64, bool) {
			if !c.StatusIsActive(skillKey) {
				return 0, false
			}
			switch ai.AttackTag {
			case attacks.AttackTagSwirlCryo:
			case attacks.AttackTagSwirlElectro:
			case attacks.AttackTagSwirlPyro:
			case attacks.AttackTagSwirlHydro:
			default:
				return 0, false
			}
			return 999 * c.Stat(attributes.EM), false
		},
	})
}
