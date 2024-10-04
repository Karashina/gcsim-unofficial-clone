package kinich

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/targets"
	"github.com/genshinsim/gcsim/pkg/enemy"
)

const (
	skillKey       = "Nightsoul's Blessing: Kinich"
	skillLinkKey   = "kinich-Link"
	particleICDKey = "kinich-particle-icd"
	blindspotKey   = "kinich-blindspot"
	c2buffKey      = "kinich-c2-buff"
)

var (
	skillFrames        []int
	skillFramesSSC     []int
	skillFramesSSCHold []int
)

func init() {
	skillFrames = frames.InitAbilSlice(46)
	skillFramesSSC = frames.InitAbilSlice(70)
	skillFramesSSCHold = frames.InitAbilSlice(60)
	attackFramesE[action.ActionAttack] = 63
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	travel, ok := p["travel"]
	if !ok {
		c.skilltravel = 10
	} else {
		c.skilltravel = travel
	}

	hold, ok := p["hold"]
	if !ok {
		c.skillhold = 0
		c.isSSCHeld = false
	} else {
		c.skillhold = 30 + hold
		c.isSSCHeld = true
	}

	if !c.StatusIsActive(skillKey) {
		return c.skillActivate(), nil
	}
	return c.skillSSC(), nil
}

func (c *char) skillActivate() action.Info {
	c.OnNightsoul = true
	c.NightsoulPoint = 0
	c.AddStatus(skillKey, 10*60, true)
	c.Core.Tasks.Add(c.generateNightsoulPoints, 43)
	c.AddStatus(blindspotKey, -1, false)
	c.SetCD(action.ActionSkill, 18*60)

	if c.Base.Cons >= 2 {
		c.AddStatus(c2buffKey, -1, false)
	}

	// Return ActionInfo
	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionSwap], // earliest cancel
		State:           action.SkillState,
	}
}

func (c *char) skillSSC() action.Info {

	SSCrelease := 42
	frame := frames.NewAbilFunc(skillFramesSSC)
	anim := skillFramesSSC[action.InvalidAction]
	cqa := skillFramesSSC[action.ActionSwap]

	if c.isSSCHeld {
		SSCrelease = c.skillhold
		frame = frames.NewAbilFunc(skillFramesSSCHold)
		anim = skillFramesSSCHold[action.InvalidAction]
		cqa = skillFramesSSCHold[action.ActionSwap]
	}

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Canopy Hunter: Riding High (Scalespiker Cannon DMG)",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagKinichSkillSSC,
		ICDGroup:   attacks.ICDGroupKinichSkillSSC,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Dendro,
		Durability: 25,
		Mult:       skillSSC[c.TalentLvlSkill()],
		FlatDmg:    c.TotalAtk() * c.a4buff,
		Alignment:  attacks.AdditionalTagNightsoul,
	}

	ap := combat.NewCircleHit(c.Core.Combat.Player(), c.Core.Combat.PrimaryTarget(), nil, c.sscradius)

	c.ConsumeNightsoul(20)
	c.AddStatus(blindspotKey, -1, false)
	c.Core.QueueAttack(ai, ap, SSCrelease+c.skillhold, SSCrelease+c.skillhold+c.skilltravel, c.particleCB, c.c2CB())

	c.c4energy()
	if c.Base.Cons >= 6 {
		c6ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Auspicious Beast's Shape (Scalespiker Cannon Bounce)",
			AttackTag:  attacks.AttackTagNone,
			ICDTag:     attacks.ICDTagKinichSkillSSC,
			ICDGroup:   attacks.ICDGroupKinichSkillSSC,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Dendro,
			Durability: 25,
			Mult:       7,
			FlatDmg:    c.TotalAtk() * c.a4buff,
			Alignment:  attacks.AdditionalTagNightsoul,
		}
		c.Core.QueueAttack(c6ai, ap, SSCrelease+c.skillhold+c.skilltravel+10, SSCrelease+c.skillhold+c.skilltravel+30, c.particleCB, c.c2CB())
	}

	if c.StatusIsActive(c2buffKey) {
		c.QueueCharTask(c.removec2, SSCrelease+c.skillhold+c.skilltravel+31)
	}

	c.a4stacks = 0
	c.a4buff = 0

	// Return ActionInfo
	return action.Info{
		Frames:          frame,
		AnimationLength: anim,
		CanQueueAfter:   cqa, // earliest cancel
		State:           action.SkillState,
	}
}

func (c *char) removec2() {
	c.DeleteStatus(c2buffKey)
}

func (c *char) skillEndRoutine() {
	c.DeleteStatus(particleICDKey)
	c.NightsoulPoint = 0
	c.OnNightsoul = false
	for _, t := range c.Core.Combat.Enemies() {
		if e, ok := t.(*enemy.Enemy); ok {
			e.DeleteStatus(A1MarkKey)
		}
	}
}

func (c *char) generateNightsoulPoints() {
	if c.StatusIsActive(skillKey) {
		c.AddNightsoul("kinich-skill-generation", 1)
		c.Core.Tasks.Add(c.generateNightsoulPoints, 30)
	}
}

func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if !c.StatusIsActive(skillKey) {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, -1, true)

	c.Core.QueueParticle(c.Base.Key.String(), 5, attributes.Dendro, c.ParticleDelay)
}

func (c *char) skillendcheck() {
	c.Core.Events.Subscribe(event.OnTick, func(args ...interface{}) bool {
		if c.StatusIsActive(skillKey) {
			return false
		}
		if !c.OnNightsoul {
			return false
		}
		// if skillkey expires & char on nightsoul - that means skill has ended
		c.skillEndRoutine()
		return false
	}, "kinich-skill-end-check")
}
