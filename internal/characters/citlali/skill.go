package citlali

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/player/shield"
	"github.com/genshinsim/gcsim/pkg/core/targets"
)

var skillFrames []int

const (
	SkillHitmark            = 23
	NightsoulDelay          = 20
	NightsoulReductionStart = 10
	ShieldDelay             = 39
	SkillCDDelay            = 20
	SkillDoTInit            = 60
	SkillDoTInterval        = 59
	SkillKey                = "citlali-skill-active"
	ItzpapaKey              = "citlali-itzpapa-active"
)

func init() {
	skillFrames = frames.InitAbilSlice(43) // E -> Q/D/J
	skillFrames[action.ActionAttack] = 41  // E -> N1
	skillFrames[action.ActionSkill] = 42   // E -> E
	skillFrames[action.ActionWalk] = 42    // E -> W
	skillFrames[action.ActionSwap] = 41    // E -> Swap
}

func (c *char) reduceNightsoulPoints(val float64) {
	consumed := val
	if c.nightsoulState.Points() <= val {
		consumed = c.nightsoulState.Points()
	}
	c.nightsoulState.ConsumePoints(consumed)
	if c.Base.Cons >= 6 {
		c.c6count += consumed
		if c.c6count >= 40 {
			c.c6count = 40
		}
		c.c6buff[attributes.DmgP] = 0.015 * c.c6count
		c.c6self[attributes.DmgP] = 0.025 * c.c6count
	}
	if c.nightsoulState.Points() <= 0.00001 && c.Base.Cons < 6 {
		c.DeleteStatus(ItzpapaKey)
	}
}

func (c *char) cancelNightsoul() {
	c.DeleteStatus(ItzpapaKey)
	c.nightsoulState.ExitBlessing()
	c.nightsoulSrc = -1
	c.c6count = 0
}

func (c *char) nightsoulPointReduceFunc(src int) func() {
	return func() {
		if c.nightsoulSrc != src {
			return
		}

		if !c.nightsoulState.HasBlessing() {
			return
		}

		if !c.StatusIsActive(ItzpapaKey) {
			return
		}

		c.reduceNightsoulPoints(0.8)

		// reduce 0.8 point per 6f
		c.QueueCharTask(c.nightsoulPointReduceFunc(src), 6)
	}
}

func (c *char) Skill(p map[string]int) (action.Info, error) {

	c.QueueCharTask(func() {
		c.nightsoulState.EnterBlessing(24)
		c.nightsoulSrc = c.Core.F
	}, NightsoulDelay)

	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Dawnfrost Darkstar: Obsidian Tzitzimitl DMG",
		AttackTag:      attacks.AttackTagElementalArt,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagNone,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Cryo,
		Durability:     25,
		Mult:           skill[c.TalentLvlSkill()],
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 6), 0, SkillHitmark, c.particleCB)
	c.AddStatus(SkillKey, 20*60, false)
	if c.Base.Cons >= 1 {
		c.SetTag(SkillKey, 10)
	}
	if c.Base.Cons >= 6 {
		c.c6()
		c.c6count = 0
		c.QueueCharTask(func() {
			c.reduceNightsoulPoints(100)
		}, NightsoulDelay+1)
	}
	c.QueueCharTask(func() {
		// add shield
		exist := c.Core.Player.Shields.Get(shield.CitlaliSkill)
		if exist == nil {
			shield := shieldperct[c.TalentLvlSkill()]*c.Stat(attributes.EM) + shieldconst[c.TalentLvlSkill()]
			c.Core.Player.Shields.Add(c.newShield(shield, 20*60))
		} else {
			shd, _ := exist.(*shd)
			shd.Expires = c.Core.F + 20*60
		}

		// apply cryo & run a task
		player, ok := c.Core.Combat.Player().(*avatar.Player)
		if !ok {
			panic("target 0 should be Player but is not!!")
		}
		player.ApplySelfInfusion(attributes.Cryo, 25, 0.1*60)
	}, ShieldDelay)

	c.SetCDWithDelay(action.ActionSkill, 16*60, SkillCDDelay)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionAttack], // earliest cancel is before skillHitmark
		State:           action.SkillState,
	}, nil
}

func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	c.Core.QueueParticle(c.Base.Key.String(), 5, attributes.Cryo, c.ParticleDelay)
}

func (c *char) SkillChecks() {
	c.Core.Events.Subscribe(event.OnTick, func(args ...interface{}) bool {

		if !c.nightsoulState.HasBlessing() {
			return false
		}

		if c.Base.Cons >= 6 && !c.StatusIsActive(ItzpapaKey) {
			c.AddStatus(ItzpapaKey, -1, false)
			c.QueueCharTask(c.nightsoulPointReduceFunc(c.nightsoulSrc), NightsoulReductionStart)
			c.QueueCharTask(c.OpalFire(c.nightsoulSrc), SkillDoTInit)
		}
		if !c.StatusIsActive(ItzpapaKey) && c.nightsoulState.Points() >= 50 {
			c.AddStatus(ItzpapaKey, -1, false)
			c.QueueCharTask(c.nightsoulPointReduceFunc(c.nightsoulSrc), NightsoulReductionStart)
			c.QueueCharTask(c.OpalFire(c.nightsoulSrc), SkillDoTInit)
		}

		if !c.StatusIsActive(SkillKey) {
			c.cancelNightsoul()
		}

		return false
	}, "citlali-skill-check")
}

func (c *char) OpalFire(src int) func() {

	return func() {

		if c.nightsoulSrc != src {
			return
		}

		if !c.nightsoulState.HasBlessing() {
			return
		}

		if !c.StatusIsActive(ItzpapaKey) {
			return
		}

		ai := combat.AttackInfo{
			ActorIndex:     c.Index,
			Abil:           "Dawnfrost Darkstar: Frostfall Storm DMG (DoT)",
			AttackTag:      attacks.AttackTagElementalArt,
			AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
			ICDTag:         attacks.ICDTagFrostfallStorm,
			ICDGroup:       attacks.ICDGroupFrostfallStorm,
			StrikeType:     attacks.StrikeTypeDefault,
			Element:        attributes.Cryo,
			Durability:     25,
			Mult:           skilldot[c.TalentLvlBurst()],
		}

		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 6), 0, 0, c.c4CB)

		c.QueueCharTask(c.OpalFire(src), SkillDoTInterval)
	}

}
