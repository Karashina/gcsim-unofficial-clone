package illuga

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
var skillHoldFrames []int

const (
	skillTapHitmark  = 27
	skillHoldHitmark = 27
)

func init() {
	skillFrames = frames.InitAbilSlice(42)

	skillHoldFrames = frames.InitAbilSlice(42)
}

// Skill performs Dawnbearing Songbird
// Throws lamp towards enemies, dealing Geo DMG based on EM and DEF
func (c *char) Skill(p map[string]int) (action.Info, error) {
	hold, ok := p["hold"]
	if !ok {
		hold = 0
	}

	if hold > 0 {
		return c.skillHold(p, hold)
	}
	return c.skillTap(p)
}

func (c *char) skillTap(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Dawnbearing Songbird (Tap)",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Geo,
		Durability: 25,
	}

	// Calculate FlatDmg from EM + DEF scaling
	em := c.Stat(attributes.EM)
	def := c.TotalDef(false)
	emMult := skillTapEM[c.TalentLvlSkill()]
	defMult := skillTapDEF[c.TalentLvlSkill()]

	ai.FlatDmg = em*emMult + def*defMult

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5)

	c.QueueCharTask(func() {
		c.Core.QueueAttack(ai, ap, 0, 0, c.particleCB)
	}, skillTapHitmark)

	// Apply A1 Lightkeeper's Oath buff
	c.applyLightkeeperOath()

	// Set cooldown (15s)
	c.SetCDWithDelay(action.ActionSkill, 15*60, skillTapHitmark)

	c.Core.Log.NewEvent("Illuga uses Dawnbearing Songbird (Tap)", glog.LogCharacterEvent, c.Index).
		Write("em", em).
		Write("def", def).
		Write("flat_dmg", ai.FlatDmg)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash],
		State:           action.SkillState,
	}, nil
}

func (c *char) skillHold(p map[string]int, hold int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Dawnbearing Songbird (Hold)",
		AttackTag:  attacks.AttackTagElementalArtHold,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Geo,
		Durability: 25,
	}

	// Calculate FlatDmg from EM + DEF scaling (Hold has higher multipliers)
	em := c.Stat(attributes.EM)
	def := c.TotalDef(false)
	emMult := skillHoldEM[c.TalentLvlSkill()]
	defMult := skillHoldDEF[c.TalentLvlSkill()]

	ai.FlatDmg = em*emMult + def*defMult

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 6)

	c.QueueCharTask(func() {
		c.Core.QueueAttack(ai, ap, 0, 0, c.particleCB)
	}, skillHoldHitmark+hold)

	// Apply A1 Lightkeeper's Oath buff
	c.applyLightkeeperOath()

	// Set cooldown (15s)
	c.SetCDWithDelay(action.ActionSkill, 15*60, skillHoldHitmark)

	c.Core.Log.NewEvent("Illuga uses Dawnbearing Songbird (Hold)", glog.LogCharacterEvent, c.Index).
		Write("em", em).
		Write("def", def).
		Write("flat_dmg", ai.FlatDmg)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillHoldFrames),
		AnimationLength: skillHoldFrames[action.InvalidAction] + hold,
		CanQueueAfter:   skillHoldFrames[action.ActionDash] + hold,
		State:           action.SkillState,
	}, nil
}

// particleCB handles particle generation
func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, 60, false)

	c.Core.QueueParticle(c.Base.Key.String(), 3, attributes.Geo, c.ParticleDelay)
}
