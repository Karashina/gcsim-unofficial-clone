package mizuki

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/core/targets"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

var skillFrames []int
var SkillChangeFrames []int

const (
	skillHitmark     = 10
	skillCDDelay     = 10
	skillDotDelay    = 57
	skillDotInterval = 40
	skillKey         = "mizuki-skill"
)

func init() {
	skillFrames = frames.InitAbilSlice(343)
	skillFrames[action.ActionCharge] = 343
	skillFrames[action.ActionSkill] = 343
	skillFrames[action.ActionDash] = 343
	skillFrames[action.ActionJump] = 343
	skillFrames[action.ActionSwap] = 343

	SkillChangeFrames = frames.InitAbilSlice(27)
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(skillKey) {
		c.DeleteStatus(skillKey)
		return action.Info{
			Frames:          frames.NewAbilFunc(SkillChangeFrames),
			AnimationLength: SkillChangeFrames[action.InvalidAction],
			CanQueueAfter:   SkillChangeFrames[action.ActionAttack], // earliest cancel
			State:           action.SkillState,
		}, nil
	}

	c.particleCount = 0

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Aisa Utamakura Pilgrimage (Initial)",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       skill[c.TalentLvlSkill()],
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 5), skillHitmark, skillHitmark, c.particleCB)

	c.AddStatus(skillKey, 5*60, false)
	c.SetTag(skillKey, 2)

	src := c.Core.F
	c.dreamdrifterSrc = src
	c.QueueCharTask(c.checkDresmdrifter(src), skillDotDelay)
	c.QueueCharTask(c.checkSnack(src), skillCDDelay)
	c.QueueCharTask(c.c1check(src), skillCDDelay)
	c.SetCDWithDelay(action.ActionSkill, 15*60, skillCDDelay)
	c.c2()

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
	if c.particleCount > 4 {
		return
	}
	c.particleCount++
	c.Core.QueueParticle(c.Base.Key.String(), 1, attributes.Anemo, c.ParticleDelay)
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
			ICDGroup:   attacks.ICDGroupDreamdrifter,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Anemo,
			Durability: 25,
			Mult:       skillDot[c.TalentLvlSkill()],
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 5), 0, 0, c.particleCB)

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
		c.QueueCharTask(c.checkSnack(src), SnackSpawnInterval)
	}
}

func (c *char) swirlBuff() {
	for _, x := range c.Core.Player.Chars() {
		char := x
		char.AddReactBonusMod(character.ReactBonusMod{
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
				c.Core.Log.NewEvent(
					"mizuki E adding swirl buff",
					glog.LogCharacterEvent,
					c.Index,
				).Write("buff: ", skillBuff[c.TalentLvlSkill()]*c.Stat(attributes.EM))

				return skillBuff[c.TalentLvlSkill()] * c.Stat(attributes.EM), false
			},
		})
	}
}
