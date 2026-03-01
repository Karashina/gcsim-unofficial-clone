package varka

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var (
	skillFrames []int
	fwaFrames   []int
	fwaHitmarks = []int{31, 52}
)

const (
	skillHitmark         = 41
	skillConversionStart = 50
)

func init() {
	skillFrames = frames.InitAbilSlice(65)

	fwaFrames = frames.InitAbilSlice(65)
	fwaFrames[action.ActionDash] = fwaHitmarks[1]
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	if c.sturmActive {
		// C6: After FWA, tap Skill triggers additional Azure Devour (not FWA)
		if c.Base.Cons >= 6 && c.StatusIsActive(c6FWAWindowKey) {
			info, err := c.azureDevour(p)
			if err == nil {
				info.State = action.SkillState
			}
			return info, err
		}
		return c.fourWindsAscension(p)
	}
	return c.windBoundExecution(p)
}

// windBoundExecution is the initial skill cast that enters Sturm und Drang
func (c *char) windBoundExecution(p map[string]int) (action.Info, error) {
	// Deal Anemo DMG (Skill DMG)
	ai := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Windbound Execution",
		AttackTag:          attacks.AttackTagElementalArt,
		ICDTag:             attacks.ICDTagNone,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeBlunt,
		PoiseDMG:           150.0,
		Element:            attributes.Anemo,
		Durability:         25,
		Mult:               skillDmg[c.TalentLvlSkill()],
		HitlagFactor:       0.01,
		HitlagHaltFrames:   0.1 * 60,
		CanBeDefenseHalted: true,
	}

	c.Core.QueueAttack(
		ai,
		combat.NewBoxHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, 4, 5),
		skillHitmark, skillHitmark,
		c.skillParticleCB,
	)

	// Enter Sturm und Drang mode
	c.enterSturmUndDrang()

	// Set CD based on hold parameter
	hold, ok := p["hold"]
	if ok && hold > 0 {
		c.SetCD(action.ActionSkill, 8*60) // Hold CD = 8s
	} else {
		c.SetCD(action.ActionSkill, 16*60) // Tap CD = 16s
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillHitmark,
		State:           action.SkillState,
	}, nil
}

// enterSturmUndDrang activates S&D mode
func (c *char) enterSturmUndDrang() {
	c.sturmActive = true
	c.sturmSrc = c.Core.F
	c.cdReductionCount = 0

	// FWA charges: starts on cooldown (0 charges), CD = 11s per charge
	c.fwaCharges = 0
	c.fwaCDEndFrame = c.Core.F + 11*60

	// S&D Duration: 12s
	dur := 12 * 60
	c.AddStatus(sturmUndDrangKey, dur, true)

	// C1: Lyrical Libation - first FWA/Azure Devour deals 200% DMG
	if c.Base.Cons >= 1 {
		c.AddStatus(c1LyricalKey, dur, true)
	}

	// Schedule exit
	src := c.sturmSrc
	c.Core.Tasks.Add(func() {
		if c.sturmSrc != src {
			return
		}
		c.exitSturmUndDrang()
	}, dur)
}

// exitSturmUndDrang deactivates S&D mode
func (c *char) exitSturmUndDrang() {
	c.sturmActive = false
	c.fwaCharges = 0
	c.DeleteStatus(sturmUndDrangKey)
	c.DeleteStatus(c1LyricalKey)
}

// fourWindsAscension handles the special skill during S&D mode
func (c *char) fourWindsAscension(p map[string]int) (action.Info, error) {
	lvl := c.TalentLvlSkill()

	// C6: check if this is from the FWA window (no charge consumption)
	consumeCharge := true
	if c.Base.Cons >= 6 {
		if c.StatusIsActive(c6AzureWindowKey) {
			// After Azure Devour, tap skill triggers additional FWA without consuming charges
			consumeCharge = false
			c.DeleteStatus(c6AzureWindowKey)
		}
		// c6FWAWindowKey case is now handled by Skill() routing to azureDevour
	}

	if consumeCharge {
		c.fwaCharges--
	}

	// C1: Lyrical Libation effect
	c1Mult := 1.0
	if c.Base.Cons >= 1 && c.StatusIsActive(c1LyricalKey) {
		c1Mult = 2.0
		c.DeleteStatus(c1LyricalKey)
	}

	// FWA: 2 hits
	// 1st: Other element (No ICD), 2nd: Anemo (No ICD)
	otherEle := c.otherElement
	if !c.hasOtherEle {
		otherEle = attributes.Anemo
	}

	// Hit 1: Other element
	mult1 := fwaOther[lvl]
	if c.Base.Ascension >= 1 && c.a1MultFactor != 1.0 {
		mult1 *= c.a1MultFactor
	}
	mult1 *= c1Mult

	ai1 := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Four Winds' Ascension (Other)",
		AttackTag:          attacks.AttackTagElementalArt,
		ICDTag:             attacks.ICDTagNone,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeBlunt,
		PoiseDMG:           150.0,
		Element:            otherEle,
		Durability:         25,
		Mult:               mult1,
		HitlagFactor:       0.01,
		HitlagHaltFrames:   0.1 * 60,
		CanBeDefenseHalted: true,
	}
	c.Core.QueueAttack(
		ai1,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, 4),
		fwaHitmarks[0], fwaHitmarks[0],
	)

	// Hit 2: Anemo
	mult2 := fwaAnemo[lvl]
	if c.Base.Ascension >= 1 && c.a1MultFactor != 1.0 {
		mult2 *= c.a1MultFactor
	}
	mult2 *= c1Mult

	ai2 := combat.AttackInfo{
		ActorIndex:         c.Index,
		Abil:               "Four Winds' Ascension (Anemo)",
		AttackTag:          attacks.AttackTagElementalArt,
		ICDTag:             attacks.ICDTagNone,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeBlunt,
		PoiseDMG:           100.0,
		Element:            attributes.Anemo,
		Durability:         25,
		Mult:               mult2,
		HitlagFactor:       0.01,
		HitlagHaltFrames:   0.08 * 60,
		CanBeDefenseHalted: true,
	}
	c.Core.QueueAttack(
		ai2,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, 4),
		fwaHitmarks[1], fwaHitmarks[1],
	)

	// C2: Additional Anemo strike equal to 800% ATK
	if c.Base.Cons >= 2 {
		c.c2Strike(fwaHitmarks[1] + 4)
	}

	// C6: After FWA, open window for additional Azure Devour
	// Only set window when this was a normal FWA (not a C6 chain trigger)
	if c.Base.Cons >= 6 && consumeCharge {
		c.AddStatus(c6FWAWindowKey, 60, true) // ~1s window
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(fwaFrames),
		AnimationLength: fwaFrames[action.InvalidAction],
		CanQueueAfter:   fwaHitmarks[0],
		State:           action.SkillState,
	}, nil
}

// skillParticleCB generates particles when skill hits enemy
func (c *char) skillParticleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, 0.5*60, true)

	// Generates 6 elemental particles
	c.Core.QueueParticle(c.Base.Key.String(), 6, attributes.Anemo, c.ParticleDelay)
}

// c2Strike performs the C2 additional Anemo strike
func (c *char) c2Strike(delay int) {
	atk := c.TotalAtk()
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "C2: Dawn's Flight Strike",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		PoiseDMG:   100.0,
		Element:    attributes.Anemo,
		Durability: 25,
		FlatDmg:    atk * 8.0, // 800% ATK
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, 5),
		delay, delay,
	)
}

func (c *char) c2Init() {
	// C2 is handled directly in FWA and Azure Devour functions
}
