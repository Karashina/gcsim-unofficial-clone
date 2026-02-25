package zibai

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var skillFrames []int
var spiritSteedFrames []int

const (
	skillHitmark        = 22
	spiritSteedHitmark1 = 31
	spiritSteedHitmark2 = 35
)

func init() {
	skillFrames = frames.InitAbilSlice(32)

	spiritSteedFrames = frames.InitAbilSlice(54)
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	// If already in Lunar Phase Shift mode and have enough radiance, use Spirit Steed's Stride
	if c.lunarPhaseShiftActive && c.phaseShiftRadiance >= spiritSteedRadianceCost {
		return c.spiritSteedStride(p)
	}

	// Enter Lunar Phase Shift mode
	c.enterLunarPhaseShift()

	c.Core.Log.NewEvent("Zibai enters Lunar Phase Shift mode", glog.LogCharacterEvent, c.Index).
		Write("duration", lunarPhaseShiftDuration)

	// Set cooldown (18s)
	c.SetCDWithDelay(action.ActionSkill, 18*60, skillHitmark)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash],
		State:           action.SkillState,
	}, nil
}

// enterLunarPhaseShift activates the Lunar Phase Shift mode
func (c *char) enterLunarPhaseShift() {
	c.lunarPhaseShiftActive = true
	src := c.Core.F
	c.lunarPhaseShiftSrc = src

	// C1: Immediately gain 100 Phase Shift Radiance
	if c.Base.Cons >= 1 {
		c.phaseShiftRadiance = 100
		c.c1FirstStride = true // First stride gets bonus
	} else {
		c.phaseShiftRadiance = 0
	}

	c.spiritSteedUsages = 0
	c.AddStatus(skillKey, lunarPhaseShiftDuration, true)

	// Start periodic radiance gain from methods
	c.startRadianceAccumulation(src)

	// Schedule mode exit (with source check so burst extension can invalidate)
	c.QueueCharTask(func() {
		if c.lunarPhaseShiftSrc != src {
			return // invalidated by extension
		}
		c.exitLunarPhaseShift()
	}, lunarPhaseShiftDuration)
}

// exitLunarPhaseShift deactivates the Lunar Phase Shift mode
func (c *char) exitLunarPhaseShift() {
	if !c.lunarPhaseShiftActive {
		return
	}
	c.lunarPhaseShiftActive = false
	c.lunarPhaseShiftSrc = -1
	c.phaseShiftRadiance = 0
	c.spiritSteedUsages = 0
	c.c1FirstStride = false
	// Reset saved normal counter if C4 is not active
	if c.Base.Cons < 4 {
		c.savedNormalCounter = 0
	}
	c.DeleteStatus(skillKey)
	c.DeleteStatus(radianceNormalICDKey)
	c.DeleteStatus(radianceLCrsICDKey)

	c.Core.Log.NewEvent("Zibai exits Lunar Phase Shift mode", glog.LogCharacterEvent, c.Index)
}

// extendLunarPhaseShift extends the duration of Lunar Phase Shift by specified frames
func (c *char) extendLunarPhaseShift(extensionFrames int) {
	if !c.lunarPhaseShiftActive {
		return
	}
	c.ExtendStatus(skillKey, extensionFrames)

	// Reschedule exit task: update source to invalidate old exit, then queue new one
	src := c.Core.F
	c.lunarPhaseShiftSrc = src
	c.startRadianceAccumulation(src)

	remaining := c.StatusDuration(skillKey)
	c.QueueCharTask(func() {
		if c.lunarPhaseShiftSrc != src {
			return // invalidated by another extension
		}
		c.exitLunarPhaseShift()
	}, remaining)

	c.Core.Log.NewEvent("Zibai Lunar Phase Shift extended", glog.LogCharacterEvent, c.Index).
		Write("extension_frames", extensionFrames).
		Write("remaining_frames", remaining)
}

// spiritSteedStride performs the Spirit Steed's Stride attack
func (c *char) spiritSteedStride(p map[string]int) (action.Info, error) {
	// C6: Consume all radiance for bonus damage
	consumedRadiance := c.phaseShiftRadiance
	var c6BonusPct float64 = 0

	if c.Base.Cons >= 6 && consumedRadiance > 70 {
		// 1.6% per point above 70
		c6BonusPct = float64(consumedRadiance-70) * 0.016
		c.applyC6ElevationBuff(c6BonusPct)
	}

	// Consume radiance (C6 consumes all, otherwise 70)
	if c.Base.Cons >= 6 {
		c.phaseShiftRadiance = 0
	} else {
		c.phaseShiftRadiance -= spiritSteedRadianceCost
	}
	c.spiritSteedUsages++
	c.DeleteStatus(radianceLCrsICDKey)

	// 1st Hit DMG
	ai1 := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Spirit Steed's Stride 1-Hit",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Geo,
		Durability: 25,
		UseDef:     true,
		Mult:       spiritSteedStride_1[c.TalentLvlSkill()],
	}

	ap1 := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5)

	c.QueueCharTask(func() {
		c.Core.QueueAttack(ai1, ap1, 0, 0, c.spiritSteedOnHitCB)
	}, spiritSteedHitmark1)

	// Calculate 2nd hit multiplier with bonuses
	secondHitMult := 1.6 * spiritSteedStride_2[c.TalentLvlSkill()]

	ai2 := combat.AttackInfo{
		ActorIndex:       c.Index,
		Abil:             "Spirit Steed's Stride 2-Hit (Lunar-Crystallize / E)",
		AttackTag:        attacks.AttackTagLCrsDamage,
		ICDTag:           attacks.ICDTagNone,
		ICDGroup:         attacks.ICDGroupDefault,
		StrikeType:       attacks.StrikeTypeDefault,
		Element:          attributes.Geo,
		Durability:       0,
		IgnoreDefPercent: 1,
	}

	// C1: First Spirit Steed's Stride 2nd-hit is increased by 220%
	c1bonus := 0.0
	if c.c1FirstStride {
		c1bonus = 2.2
		c.c1FirstStride = false
	}

	// HP scaling with Lunar-Crystallize formula
	em := c.Stat(attributes.EM)
	baseDmg := c.TotalDef(false) * secondHitMult
	emBonus := (6 * em) / (2000 + em)
	ai2.FlatDmg = baseDmg * (1 + c.LCrsBaseReactBonus(ai2) + c1bonus) * (1 + emBonus + c.LCrsReactBonus(ai2))

	// A1: Selenic Descent effect - increases 2nd hit by 60% DEF
	// C2: When Moonsign is Ascendant Gleam, A1 is further increased by 550% DEF
	// (Additional flat damage, handled in asc.go)
	if c.StatusIsActive(selenicDescentKey) {
		if c.Base.Cons >= 2 && c.isMoonsignAscendant() {
			ai2.FlatDmg += 5.5 * c.TotalDef(false)
		} else {
			ai2.FlatDmg += 0.6 * c.TotalDef(false)
		}
	}

	ai2.FlatDmg *= (1 + c.ElevationBonus(ai2))

	snap := combat.Snapshot{CharLvl: c.Base.Level}
	snap.Stats[attributes.CR] = c.Stat(attributes.CR)
	snap.Stats[attributes.CD] = c.Stat(attributes.CD)

	ap2 := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 8)
	c.QueueCharTask(func() {
		c.Core.QueueAttackWithSnap(ai2, snap, ap2, 0, c.spiritSteedOnHitCB)
	}, spiritSteedHitmark2)

	c.Core.Log.NewEvent("Zibai uses Spirit Steed's Stride", glog.LogCharacterEvent, c.Index).
		Write("radiance_consumed", consumedRadiance).
		Write("usages", c.spiritSteedUsages).
		Write("c6_bonus_pct", c6BonusPct)

	// Spec: Exit Lunar Phase Shift after 4 uses (C0) or max uses (C1: 5)
	if c.spiritSteedUsages >= c.maxSpiritSteedUsages {
		c.QueueCharTask(func() {
			c.exitLunarPhaseShift()
		}, spiritSteedHitmark2+1) // Exit after 2nd hit lands
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(spiritSteedFrames),
		AnimationLength: spiritSteedFrames[action.InvalidAction],
		CanQueueAfter:   spiritSteedFrames[action.ActionDash],
		State:           action.SkillState,
	}, nil
}

// spiritSteedOnHitCB callback for Spirit Steed's Stride hit
func (c *char) spiritSteedOnHitCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	// C4: Gain Scattermoon Splendor effect
	if c.Base.Cons >= 4 {
		c.c4ScattermoonUsed = true
	}
}

func (c *char) initRadianceHandlers() {
	c.Core.Events.Subscribe(event.OnLunarCrystallize, func(args ...interface{}) bool {
		if !c.lunarPhaseShiftActive {
			return false
		}
		if c.StatusIsActive(radianceLCrsICDKey) {
			return false
		}
		c.AddStatus(radianceLCrsICDKey, radianceLCrsICD, false)
		c.addPhaseShiftRadiance(radianceLCrsGain)
		return false
	}, "zibai-radiance-lcrs")
}

// startRadianceAccumulation starts periodic radiance accumulation
func (c *char) startRadianceAccumulation(src int) {
	// Gain radiance over time
	c.QueueCharTask(func() {
		if c.lunarPhaseShiftSrc != src {
			return
		}
		if !c.lunarPhaseShiftActive {
			return
		}
		c.addPhaseShiftRadiance(radianceTickGain)
		c.startRadianceAccumulation(src)
	}, radianceTickInterval)
}

// particleCB handles particle generation
func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, 2*60, true)
	c.Core.QueueParticle(c.Base.Key.String(), 1, attributes.Geo, c.ParticleDelay)
}
