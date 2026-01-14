package durin

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

// Frame data: Animation lock and cancellable timing for each skill action
var (
	skillEssentialFrames []int // Frame data for Essential Transmutation
	skillFrames          []int // Frame data for Confirmation of Purity
	skillDenialFrames    []int // Frame data for Denial of Darkness
)

// Timing constants: Hitmarks, cooldowns, durations
const (
	skillHitmark        = 33      // Hitmark for Confirmation of Purity (frames)
	skillDenialHitmark1 = 15      // Hitmark for Denial of Darkness 1st hit
	skillDenialHitmark2 = 25      // Hitmark for Denial of Darkness 2nd hit
	skillDenialHitmark3 = 35      // Hitmark for Denial of Darkness 3rd hit
	skillCD             = 12 * 60 // Skill cooldown: 12 seconds
	stateDuration       = 30 * 60 // Transmutation state duration: 30 seconds
	energyRegenICD      = 6 * 60  // Energy restoration internal cooldown: 6 seconds
	skillParticleCount  = 4       // Elemental particle count
)

// State management keys: Status identifiers used internally in the simulator
const (
	essentialTransmutationKey = "durin-essential-transmutation" // Essential Transmutation state
	confirmationStateKey      = "durin-confirmation-state"      // Confirmation of Purity state
	denialStateKey            = "durin-denial-state"            // Denial of Darkness state
	skillRecastCDKey          = "durin-skill-recast-cd"         // CD for second consecutive skill use
)

func init() {
	// Essential Transmutation: E->NA: 16, E->E: 19
	skillEssentialFrames = frames.InitAbilSlice(50)
	skillEssentialFrames[action.ActionAttack] = 16
	skillEssentialFrames[action.ActionSkill] = 19
	skillEssentialFrames[action.ActionDash] = 16

	// Confirmation of Purity: Confirmation of Purity -> NA: 64
	skillFrames = frames.InitAbilSlice(64)
	skillFrames[action.ActionAttack] = 64
	skillFrames[action.ActionBurst] = 35
	skillFrames[action.ActionDash] = 30
	skillFrames[action.ActionJump] = 30
	skillFrames[action.ActionSwap] = 40

	// Denial of Darkness: Denial of Darkness -> NA: 62
	skillDenialFrames = frames.InitAbilSlice(62)
	skillDenialFrames[action.ActionAttack] = 62
	skillDenialFrames[action.ActionBurst] = 45
	skillDenialFrames[action.ActionDash] = 40
	skillDenialFrames[action.ActionJump] = 40
	skillDenialFrames[action.ActionSwap] = 50
}

// Skill is the entry point for Elemental Skill
// Branches to Confirmation of Purity or Denial of Darkness based on Essential Transmutation state
func (c *char) Skill(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(essentialTransmutationKey) {
		c.particleIcd = false         // Reset particle ICD for new skill use
		return c.skillConfirmation(p) // During Essential Transmutation → Confirmation of Purity
	}
	return c.skillEssentialTransmutation(p) // Normal state → Enter Essential Transmutation
}

// skillEssentialTransmutation enters Essential Transmutation state
// Next skill use will enable Confirmation of Purity or Denial of Darkness
func (c *char) skillEssentialTransmutation(p map[string]int) (action.Info, error) {
	c.AddStatus(essentialTransmutationKey, stateDuration, true)
	c.stateDenial = false
	c.DeleteStatus(denialStateKey)

	c.Core.Log.NewEvent("Durin enters Essential Transmutation", glog.LogCharacterEvent, c.Index)
	c.SetCDWithDelay(action.ActionSkill, skillCD, 0)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillEssentialFrames),
		AnimationLength: skillEssentialFrames[action.InvalidAction],
		CanQueueAfter:   skillEssentialFrames[action.ActionDash],
		State:           action.SkillState,
	}, nil
}

// skillConfirmation executes Confirmation of Purity: AoE Pyro DMG
// Triggered when using skill again after entering Essential Transmutation via normal attack
func (c *char) skillConfirmation(p map[string]int) (action.Info, error) {
	// GU: 1U (Durability: 25)
	// ICD Tag: None (ICDTagNone)
	// ICD Group: None - ICDGroup is ignored when using ICDTagNone
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Transmutation: Confirmation of Purity",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		Element:    attributes.Pyro,
		Durability: 25,
		Mult:       skillPurity[c.TalentLvlSkill()],
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0}, 3.0),
		skillHitmark,
		skillHitmark,
		c.skillParticleCB,
		c.skillEnergyRegenCB,
	)

	c.transitionToConfirmationState()
	c.AddStatus(skillRecastCDKey, 0, true) // Mark recast as used

	c.Core.Log.NewEvent("Durin uses Confirmation of Purity", glog.LogCharacterEvent, c.Index)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash],
		State:           action.SkillState,
	}, nil
}

// skillDenialOfDarkness executes Denial of Darkness: 3 consecutive single-target Pyro DMG hits
// Triggered when using skill again after entering Essential Transmutation via charged attack
func (c *char) skillDenialOfDarkness(p map[string]int) (action.Info, error) {
	hitmarks := []int{skillDenialHitmark1, skillDenialHitmark2, skillDenialHitmark3}
	mults := [][]float64{skillDenial1, skillDenial2, skillDenial3}
	ap := combat.NewSingleTargetHit(c.Core.Combat.PrimaryTarget().Key())

	for i := 0; i < 3; i++ {
		ai := c.makeSkillDenialAttackInfo(i+1, mults[i])
		c.Core.QueueAttack(ai, ap, hitmarks[i], hitmarks[i], c.skillParticleCB, c.skillEnergyRegenCB)
	}

	c.transitionToDenialState()
	c.AddStatus(skillRecastCDKey, 0, true) // Mark recast as used
	c.Core.Log.NewEvent("Durin uses Denial of Darkness", glog.LogCharacterEvent, c.Index)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillDenialFrames),
		AnimationLength: skillDenialFrames[action.InvalidAction],
		CanQueueAfter:   skillDenialFrames[action.ActionDash],
		State:           action.SkillState,
	}, nil
}

// makeSkillDenialAttackInfo generates attack info for each hit of Denial of Darkness
// Called for each of the 3 hits
func (c *char) makeSkillDenialAttackInfo(hitNum int, mult []float64) combat.AttackInfo {
	// GU: 1U (Durability: 25)
	// ICD Tag: Elemental Skill (ICDTagElementalArt)
	// ICD Group: Standard (ICDGroupDefault)
	return combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Transmutation: Denial of Darkness (Hit " + string(rune('0'+hitNum)) + ")",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagElementalArt,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeSlash,
		Element:    attributes.Pyro,
		Durability: 25,
		Mult:       mult[c.TalentLvlSkill()],
	}
}

// State transition helper functions

// transitionToConfirmationState transitions to Confirmation of Purity state
// This state persists for 30 seconds after using Confirmation of Purity
func (c *char) transitionToConfirmationState() {
	c.DeleteStatus(essentialTransmutationKey)
	c.stateDenial = false
	c.DeleteStatus(denialStateKey)
	c.AddStatus(confirmationStateKey, stateDuration, true)
}

// transitionToDenialState transitions to Denial of Darkness state
// This state persists for 30 seconds after using Denial of Darkness, then reverts to Confirmation of Purity
func (c *char) transitionToDenialState() {
	c.DeleteStatus(essentialTransmutationKey)
	c.stateDenial = true
	c.AddStatus(denialStateKey, stateDuration, true)

	// Schedule revert to Confirmation after duration
	c.Core.Tasks.Add(func() {
		if c.StatusIsActive(denialStateKey) {
			c.stateDenial = false
			c.DeleteStatus(denialStateKey)
			c.AddStatus(confirmationStateKey, -1, true)
			c.Core.Log.NewEvent("Durin reverts to Confirmation of Purity state", glog.LogCharacterEvent, c.Index)
		}
	}, stateDuration)
}

// Callback functions: Additional effects on attack hit

// skillParticleCB generates elemental particles on skill hit
func (c *char) skillParticleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.particleIcd {
		return
	}
	c.particleIcd = true
	c.Core.QueueParticle(c.Base.Key.String(), skillParticleCount, attributes.Pyro, c.ParticleDelay)
}

// skillEnergyRegenCB restores elemental energy on skill hit
// Has a 6 second internal cooldown
func (c *char) skillEnergyRegenCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	// Check 6 second ICD
	if c.Core.F-c.lastEnergyRestoreFrame < energyRegenICD {
		return
	}

	c.lastEnergyRestoreFrame = c.Core.F
	energyRegen := skillEnergyRegen[c.TalentLvlSkill()]
	c.AddEnergy("durin-skill-energy", energyRegen)

	c.Core.Log.NewEvent("Durin restores energy from skill", glog.LogCharacterEvent, c.Index).
		Write("energy", energyRegen)
}
