package linnea

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
var skillMashFrames []int

const (
	skillHitmark     = 17  // tE->D: 17 (CD start frame)
	skillMashHitmark = 111 // mE->hitmark: 111 (Million Ton Crush hitmark)
)

func init() {
	skillFrames = frames.InitAbilSlice(17)     // tE->D: 17
	skillMashFrames = frames.InitAbilSlice(97) // mE->D: 97
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	// mash=1: transition to Ultimate Power Form
	if p["mash"] == 1 {
		return c.skillMash(p)
	}

	// tap: summon Lumi and enter Super Power Form
	c.summonLumi(lumiFormSuper, lumiFirstTickFromE)

	// C1: add Field Catalog stacks
	if c.Base.Cons >= 1 {
		c.c1OnSkillUse()
	}

	// increase poise
	c.AddStatus("linnea-poise", lumiDuration, true)

	// set cooldown
	c.SetCDWithDelay(action.ActionSkill, skillCD, skillHitmark)

	c.Core.Log.NewEvent("Linnea summons Lumi in Super Power Form", glog.LogCharacterEvent, c.Index).
		Write("form", "super").
		Write("duration", lumiDuration)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash],
		State:           action.SkillState,
	}, nil
}

// skillMash transitions Lumi to Ultimate Power Form and triggers Million Ton Crush
func (c *char) skillMash(p map[string]int) (action.Info, error) {
	// Million Ton Crush damage (Lunar-Crystallize reaction damage)
	ai := combat.AttackInfo{
		ActorIndex:       c.Index,
		Abil:             "Lumi Million Ton Crush (Lunar-Crystallize)",
		AttackTag:        attacks.AttackTagLCrsDamage,
		ICDTag:           attacks.ICDTagNone,
		ICDGroup:         attacks.ICDGroupDefault,
		StrikeType:       attacks.StrikeTypeBlunt,
		Element:          attributes.Geo,
		Durability:       0,
		IgnoreDefPercent: 1,
	}

	// DEF scaling via Lunar-Crystallize formula
	mult := skillMillionTonCrush[c.TalentLvlSkill()]
	em := c.Stat(attributes.EM)
	baseDmg := c.TotalDef(false) * mult
	emBonus := (6 * em) / (2000 + em)
	ai.FlatDmg = baseDmg * (1 + c.LCrsBaseReactBonus(ai)) * (1 + emBonus + c.LCrsReactBonus(ai))
	ai.FlatDmg *= (1 + c.ElevationBonus(ai))

	// C1: consume Field Catalog stacks for Million Ton Crush
	if c.Base.Cons >= 1 {
		ai.FlatDmg += c.c1MillionTonCrushBonus()
	}

	// C2: CRIT DMG bonus for Million Ton Crush
	snap := combat.Snapshot{CharLvl: c.Base.Level}
	snap.Stats[attributes.CR] = c.Stat(attributes.CR)
	snap.Stats[attributes.CD] = c.Stat(attributes.CD)
	if c.Base.Cons >= 2 {
		snap.Stats[attributes.CD] += 1.50
	}

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 8)

	c.QueueCharTask(func() {
		c.Core.QueueAttackWithSnap(ai, snap, ap, 0, c.particleCB)
	}, skillMashHitmark)

	// transition to Standard Power Form
	c.lumiForm = lumiFormStandard
	c.lumiComboIdx = 0
	// first tick: mE hitmark(111) + hitmark->PPP1(132) = 243f
	src := c.Core.F
	c.lumiTickSrc = src
	c.QueueCharTask(func() {
		if c.lumiTickSrc != src {
			return
		}
		if !c.lumiActive {
			return
		}
		c.lumiAttackTick()
		c.startLumiTicks(src)
	}, lumiStdFirstTickAfterMash)

	c.Core.Log.NewEvent("Lumi uses Million Ton Crush, switching to Standard Form",
		glog.LogCharacterEvent, c.Index)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillMashFrames),
		AnimationLength: skillMashFrames[action.InvalidAction],
		CanQueueAfter:   skillMashFrames[action.ActionDash],
		State:           action.SkillState,
	}, nil
}

// summonLumi summons Lumi. initialDelay is the frame delay until the first attack tick
func (c *char) summonLumi(form lumiForm, initialDelay int) {
	c.lumiActive = true
	c.lumiSrc = c.Core.F
	c.lumiForm = form
	c.lumiComboIdx = 0

	c.AddStatus(lumiKey, lumiDuration, true)

	// first tick fires after initialDelay, then continues at normal intervals
	src := c.Core.F
	c.lumiTickSrc = src
	c.QueueCharTask(func() {
		if c.lumiTickSrc != src {
			return
		}
		if !c.lumiActive {
			return
		}
		c.lumiAttackTick()
		c.startLumiTicks(src)
	}, initialDelay)

	// dismiss Lumi when duration expires
	lumiSrc := c.lumiSrc
	c.QueueCharTask(func() {
		if c.lumiSrc != lumiSrc {
			return // invalidated by reset
		}
		c.dismissLumi()
	}, lumiDuration)
}

// dismissLumi dismisses Lumi
func (c *char) dismissLumi() {
	if !c.lumiActive {
		return
	}
	c.lumiActive = false
	c.lumiSrc = -1
	c.lumiForm = lumiFormNone
	c.lumiTickSrc = -1
	c.lumiComboIdx = 0
	c.DeleteStatus(lumiKey)

	c.Core.Log.NewEvent("Lumi dismissed", glog.LogCharacterEvent, c.Index)
}

// resetLumiDuration resets Lumi's duration without changing form
func (c *char) resetLumiDuration() {
	if !c.lumiActive {
		return
	}
	src := c.Core.F
	c.lumiSrc = src
	c.AddStatus(lumiKey, lumiDuration, true)

	// reschedule dismiss task
	c.QueueCharTask(func() {
		if c.lumiSrc != src {
			return
		}
		c.dismissLumi()
	}, lumiDuration)

	c.Core.Log.NewEvent("Lumi duration reset", glog.LogCharacterEvent, c.Index).
		Write("form", c.lumiForm)
}

// startLumiTicks starts Lumi's periodic attack ticks
func (c *char) startLumiTicks(src int) {
	tickRate := c.nextLumiTickRate()

	c.QueueCharTask(func() {
		if c.lumiTickSrc != src {
			return // invalidated
		}
		if !c.lumiActive {
			return
		}

		c.lumiAttackTick()

		// schedule next tick
		c.startLumiTicks(src)
	}, tickRate)
}

// nextLumiTickRate returns the frame interval until the next attack tick.
// comboIdx references the value already updated by the preceding lumiAttackTick call.
func (c *char) nextLumiTickRate() int {
	if c.lumiForm == lumiFormStandard {
		return lumiStandardTickRate
	}
	// with Moondrifts, interval varies based on combo position
	if c.MoonsignAscendant {
		switch c.lumiComboIdx {
		case 0:
			// after HOH -> next PPP: 61f
			return lumiSuperHOHToPPP
		case 2:
			// after 2nd PPP -> next HOH: 109f
			return lumiSuperPPPToHOH
		}
	}
	// PPP->PPP (no Moondrifts or after 1st PPP): 141f
	return lumiSuperTickRate
}

// lumiAttackTick executes one Lumi attack tick
func (c *char) lumiAttackTick() {
	switch c.lumiForm {
	case lumiFormSuper:
		c.lumiSuperFormAttack()
	case lumiFormStandard:
		c.lumiStandardFormAttack()
	default:
		// Ultimate Form is handled directly in skillMash
		return
	}
}

// lumiSuperFormAttack executes the Super Power Form attack.
// with Moondrifts: punch x2 -> hammer x1 cycle
// without Moondrifts: punch x2 only
func (c *char) lumiSuperFormAttack() {
	hasMoondrifts := c.MoonsignAscendant

	if hasMoondrifts && c.lumiComboIdx == 2 {
		// Heavy Overdrive Hammer (Lunar-Crystallize reaction damage)
		c.lumiHeavyOverdriveHammer()
		c.lumiComboIdx = 0
	} else {
		// Pound-Pound Pummeler (2-hit Geo DMG)
		c.lumiPoundPoundPummeler()
		c.lumiComboIdx++
		if !hasMoondrifts {
			c.lumiComboIdx = 0 // reset when no Moondrifts
		}
	}
}

// lumiStandardFormAttack executes the Standard Power Form attack
func (c *char) lumiStandardFormAttack() {
	c.lumiPoundPoundPummeler()
}

// lumiPoundPoundPummeler executes the Pound-Pound Pummeler attack (2 hits)
func (c *char) lumiPoundPoundPummeler() {
	for i := 0; i < 2; i++ {
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Lumi Pound-Pound Pummeler",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagElementalArt,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeBlunt,
			Element:    attributes.Geo,
			Durability: 25,
			UseDef:     true,
			Mult:       skillPoundPound[c.TalentLvlSkill()],
		}

		delay := i * 21 // inter-hit interval: 21f
		ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5)
		c.QueueCharTask(func() {
			c.Core.QueueAttack(ai, ap, 0, 0, c.particleCB)
		}, delay)
	}
}

// lumiHeavyOverdriveHammer executes the Heavy Overdrive Hammer attack (Lunar-Crystallize reaction damage)
func (c *char) lumiHeavyOverdriveHammer() {
	ai := combat.AttackInfo{
		ActorIndex:       c.Index,
		Abil:             "Lumi Heavy Overdrive Hammer (Lunar-Crystallize)",
		AttackTag:        attacks.AttackTagLCrsDamage,
		ICDTag:           attacks.ICDTagNone,
		ICDGroup:         attacks.ICDGroupDefault,
		StrikeType:       attacks.StrikeTypeBlunt,
		Element:          attributes.Geo,
		Durability:       0,
		IgnoreDefPercent: 1,
	}

	// DEF scaling via Lunar-Crystallize formula
	mult := skillHeavyOverdrive[c.TalentLvlSkill()]
	em := c.Stat(attributes.EM)
	baseDmg := c.TotalDef(false) * mult
	emBonus := (6 * em) / (2000 + em)
	ai.FlatDmg = baseDmg * (1 + c.LCrsBaseReactBonus(ai)) * (1 + emBonus + c.LCrsReactBonus(ai))
	ai.FlatDmg *= (1 + c.ElevationBonus(ai))
	ai.FlatDmg += c.LCrsFlatBonus(ai)

	snap := combat.Snapshot{CharLvl: c.Base.Level}
	snap.Stats[attributes.CR] = c.Stat(attributes.CR)
	snap.Stats[attributes.CD] = c.Stat(attributes.CD)

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 6)
	c.Core.QueueAttackWithSnap(ai, snap, ap, 0)
}

// particleCB is the particle generation callback (with ICD)
func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, 9*60, true)

	count := 3.0
	c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Geo, c.ParticleDelay)
}
