package durin

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
)

// Frame data: Animation lock and cancellable timing for each elemental burst
var (
	burstFrames         []int // Frame data for Principle of Purity
	burstFramesDarkness []int // Frame data for Principle of Darkness
)

// Timing constants - Principle of Purity
const (
	burstPurityHitmark1       = 98  // 1st hit hitmark
	burstPurityHitmark2       = 122 // 2nd hit hitmark
	burstPurityHitmark3       = 148 // 3rd hit hitmark
	burstPurityFirstDragonHit = 157 // Dragon of White Flame first hit
	dragonWhiteFlameInterval  = 59  // Dragon of White Flame attack interval
)

// Timing constants - Principle of Darkness
const (
	burstDarknessHitmark1       = 87  // 1st hit hitmark
	burstDarknessHitmark2       = 128 // 2nd hit hitmark
	burstDarknessHitmark3       = 154 // 3rd hit hitmark
	burstDarknessFirstDragonHit = 175 // Dragon of Dark Decay first hit
	dragonDarkDecayInterval     = 74  // Dragon of Dark Decay attack interval
)

// Common elemental burst constants
const (
	burstCD        = 18 * 60 // Elemental Burst cooldown: 18 seconds
	dragonDuration = 20 * 60 // Dragon duration: 20 seconds
)

// Dragon state keys: Status identifiers used internally in the simulator
const (
	dragonWhiteFlameKey = "durin-dragon-white-flame" // Dragon of White Flame state
	dragonDarkDecayKey  = "durin-dragon-dark-decay"  // Dragon of Dark Decay state
)

func init() {
	// Principle of Purity frames
	burstFrames = frames.InitAbilSlice(122)

	// Principle of Darkness frames
	burstFramesDarkness = frames.InitAbilSlice(103)
}

// Burst is the entry point for Elemental Burst
// Triggers Principle of Purity or Principle of Darkness based on transmutation state
func (c *char) Burst(p map[string]int) (action.Info, error) {
	if c.stateDenial {
		return c.burstDarkness(p) // Denial of Darkness state → Principle of Darkness
	}
	return c.burstPurity(p) // Confirmation of Purity state → Principle of Purity
}

// makeBurstAttackInfo is a helper function that generates attack info for elemental burst
func (c *char) makeBurstAttackInfo(abilName string, mult []float64) combat.AttackInfo {
	// GU: 1U (Durability: 25)
	// ICD Tag: Elemental Burst (ICDTagElementalBurst)
	// ICD Group: Standard (ICDGroupDefault)
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       abilName,
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		Element:    attributes.Pyro,
		Durability: 25,
		Mult:       mult[c.TalentLvlBurst()],
	}

	return ai
}

// makeBurstDarknessAttackInfo is a helper function that generates attack info for Principle of Darkness (Elemental Burst)
func (c *char) makeBurstDarknessAttackInfo(abilName string, mult []float64) combat.AttackInfo {
	// GU: 1U (Durability: 25)
	// ICD Tag: Elemental Burst (ICDTagElementalBurst)
	// ICD Group: Standard (ICDGroupDefault)
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       abilName,
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		Element:    attributes.Pyro,
		Durability: 25,
		Mult:       mult[c.TalentLvlBurst()],
	}

	return ai
}

// applyC6DefIgnore is a helper function that applies C6 Constellation DEF ignore effect
// Principle of Purity: 30% DEF ignore (additionally Dragon of White Flame applies DEF shred on hit - implemented in cons.go)
// Principle of Darkness: 70% DEF ignore (30% base + 40% additional)
func (c *char) applyC6DefIgnore(attacks []*combat.AttackInfo, isDarkness bool) {
	if c.Base.Cons < 6 {
		return
	}
	defIgnore := 0.3
	if isDarkness {
		defIgnore = 0.7 // 30% base + 40% additional
	}
	for _, ai := range attacks {
		ai.IgnoreDefPercent = defIgnore
	}
}

// applyBurstEffects is a helper function that applies common elemental burst effects
// Handles A4 talent, C2 constellation, energy consumption, and cooldown setting
func (c *char) applyBurstEffects() {
	// A4: Gain Primordial Fusion stacks (10 stacks, 20 seconds)
	c.a4OnBurst()

	// C2: Enable elemental reaction buff window (20 seconds)
	if c.Base.Cons >= 2 {
		c.AddStatus(c2BuffKey, c2BuffDuration, true)
	}

	// Consume elemental energy and set cooldown
	c.ConsumeEnergy(4)
	c.SetCDWithDelay(action.ActionBurst, burstCD, 2)
}

// clearExistingDragons clears any existing dragons when a new burst is used
// This prevents overlapping dragon effects
func (c *char) clearExistingDragons() {
	// Clear dragon flags
	c.dragonWhiteFlame = false
	c.dragonDarkDecay = false
	c.dragonExpiry = 0
	c.dragonSrc++ // Increment source ID to invalidate old dragon tasks

	// Delete status keys (will prevent scheduled attacks from executing)
	c.DeleteStatus(dragonWhiteFlameKey)
	c.DeleteStatus(dragonDarkDecayKey)

	c.Core.Log.NewEvent("Cleared existing dragons", glog.LogCharacterEvent, c.Index).
		Write("new_dragon_src", c.dragonSrc)
}

// burstPurity executes Principle of Purity: As the Light Shifts
// 3 instances of AoE Pyro DMG + summon Dragon of White Flame (20 seconds duration, attacks every 59 frames in AoE)
func (c *char) burstPurity(p map[string]int) (action.Info, error) {
	// Clear any existing dragons before summoning new one
	c.clearExistingDragons()

	// Create 3 attack instances
	ai1 := c.makeBurstAttackInfo("Principle of Purity: As the Light Shifts (Hit 1)", burstPurity1)
	ai2 := c.makeBurstAttackInfo("Principle of Purity: As the Light Shifts (Hit 2)", burstPurity2)
	ai3 := c.makeBurstAttackInfo("Principle of Purity: As the Light Shifts (Hit 3)", burstPurity3)

	// C6: Apply DEF ignore
	c.applyC6DefIgnore([]*combat.AttackInfo{&ai1, &ai2, &ai3}, false)

	// Queue attacks
	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 2}, 5.0)
	c.Core.QueueAttack(ai1, ap, burstPurityHitmark1, burstPurityHitmark1)
	c.Core.QueueAttack(ai2, ap, burstPurityHitmark2, burstPurityHitmark2)
	c.Core.QueueAttack(ai3, ap, burstPurityHitmark3, burstPurityHitmark3)

	// Summon dragon and apply effects
	c.summonDragonWhiteFlame()
	c.applyBurstEffects()

	// C1: Apply Cycle of Enlightenment to other party members
	if c.Base.Cons >= 1 {
		c.c1OnBurstPurity()
	}

	c.Core.Log.NewEvent("Durin uses Principle of Purity: As the Light Shifts", glog.LogCharacterEvent, c.Index)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash],
		State:           action.BurstState,
	}, nil
}

// burstDarkness executes Principle of Darkness: As the Stars Smolder
// 3 instances of AoE Pyro DMG + summon Dragon of Dark Decay (20 seconds duration, attacks every 74 frames in single-target)
func (c *char) burstDarkness(p map[string]int) (action.Info, error) {
	// Clear any existing dragons before summoning new one
	c.clearExistingDragons()

	// Create 3 attack instances
	ai1 := c.makeBurstDarknessAttackInfo("Principle of Darkness: As the Stars Smolder (Hit 1)", burstDarkness1)
	ai2 := c.makeBurstDarknessAttackInfo("Principle of Darkness: As the Stars Smolder (Hit 2)", burstDarkness2)
	ai3 := c.makeBurstDarknessAttackInfo("Principle of Darkness: As the Stars Smolder (Hit 3)", burstDarkness3)

	// C6: Apply DEF ignore (70% for Darkness)
	c.applyC6DefIgnore([]*combat.AttackInfo{&ai1, &ai2, &ai3}, true)

	// Queue attacks
	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 2}, 5.0)
	c.Core.QueueAttack(ai1, ap, burstDarknessHitmark1, burstDarknessHitmark1)
	c.Core.QueueAttack(ai2, ap, burstDarknessHitmark2, burstDarknessHitmark2)
	c.Core.QueueAttack(ai3, ap, burstDarknessHitmark3, burstDarknessHitmark3)

	// Summon dragon and apply effects
	c.summonDragonDarkDecay()
	c.applyBurstEffects()

	// C1: Apply Cycle of Enlightenment to Durin
	if c.Base.Cons >= 1 {
		c.c1OnBurstDarkness()
	}

	c.Core.Log.NewEvent("Durin uses Principle of Darkness: As the Stars Smolder", glog.LogCharacterEvent, c.Index)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFramesDarkness),
		AnimationLength: burstFramesDarkness[action.InvalidAction],
		CanQueueAfter:   burstFramesDarkness[action.ActionDash],
		State:           action.BurstState,
	}, nil
}

// makeDragonAttackInfo is a helper function that generates attack info for dragon attacks
// C6: Dragon of White Flame has 30% DEF ignore and applies DEF shred on hit (handled via callback)
// C6: Dragon of Dark Decay has 70% DEF ignore (30% base + 40% additional)
func (c *char) makeDragonAttackInfo(abilName string, mult []float64, isDarkness bool) combat.AttackInfo {
	// GU: 1U (Durability: 25)
	// ICD Tag: Durin DoT (ICDTagDurinDoT)
	// ICD Group: Time-based (White Flame 90f, Dark Decay 120f)
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       abilName,
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagDurinDoT,
		ICDGroup:   attacks.ICDGroupDurinDoTWhiteFlame,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Pyro,
		Durability: 25,
		Mult:       mult[c.TalentLvlBurst()],
	}

	// Dragon of Dark Decay uses 120f ICD group
	if isDarkness {
		ai.ICDGroup = attacks.ICDGroupDurinDoTDarkDecay
	}

	// C6: DEF ignore
	if c.Base.Cons >= 6 {
		if isDarkness {
			ai.IgnoreDefPercent = 0.7 // 30% base + 40% additional
		} else {
			ai.IgnoreDefPercent = 0.3 // Dragon also applies DEF shred on hit via callback
		}
	}

	return ai
}

// summonDragonWhiteFlame summons Dragon of White Flame
// Lasts 20 seconds, deals AoE Pyro DMG every 59 frames
func (c *char) summonDragonWhiteFlame() {
	c.dragonDarkDecay = false
	c.dragonWhiteFlame = true
	c.dragonExpiry = c.Core.F + dragonDuration

	c.AddStatus(dragonWhiteFlameKey, dragonDuration, true)
	c.DeleteStatus(dragonDarkDecayKey)

	// Capture current source ID for this dragon instance
	dragonSrc := c.dragonSrc

	c.Core.Log.NewEvent("Dragon of White Flame summoned", glog.LogCharacterEvent, c.Index).
		Write("dragon_src", dragonSrc)

	// 定期攻撃を開始
	c.Core.Tasks.Add(func() {
		c.dragonWhiteFlameAttack(0, dragonSrc)
	}, burstPurityFirstDragonHit)
}

// dragonWhiteFlameAttack executes periodic attacks of Dragon of White Flame
// Automatically schedules next attack while duration is active
func (c *char) dragonWhiteFlameAttack(attackNum int, src int) {
	// Check if this dragon instance is still valid
	if src != c.dragonSrc {
		return
	}
	if !c.StatusIsActive(dragonWhiteFlameKey) {
		return
	}

	ai := c.makeDragonAttackInfo("Dragon of White Flame", dragonWhiteFlameDmg, false)

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3.0),
		0,
		0,
		c.a4DragonAttackCB,
		c.c6DragonWhiteFlameCB,
	)

	c.Core.Tasks.Add(func() {
		c.dragonWhiteFlameAttack(attackNum+1, src)
	}, dragonWhiteFlameInterval)
}

// summonDragonDarkDecay summons Dragon of Dark Decay
// Lasts 20 seconds, deals single-target Pyro DMG every 74 frames
func (c *char) summonDragonDarkDecay() {
	c.dragonWhiteFlame = false
	c.dragonDarkDecay = true
	c.dragonExpiry = c.Core.F + dragonDuration

	c.AddStatus(dragonDarkDecayKey, dragonDuration, true)
	c.DeleteStatus(dragonWhiteFlameKey)

	// Capture current source ID for this dragon instance
	dragonSrc := c.dragonSrc

	c.Core.Log.NewEvent("Dragon of Dark Decay summoned", glog.LogCharacterEvent, c.Index).
		Write("dragon_src", dragonSrc)

	// 定期攻撃を開始
	c.Core.Tasks.Add(func() {
		c.dragonDarkDecayAttack(0, dragonSrc)
	}, burstDarknessFirstDragonHit)
}

// dragonDarkDecayAttack executes periodic attacks of Dragon of Dark Decay
// Automatically schedules next attack while duration is active
func (c *char) dragonDarkDecayAttack(attackNum int, src int) {
	// Check if this dragon instance is still valid
	if src != c.dragonSrc {
		return
	}
	if !c.StatusIsActive(dragonDarkDecayKey) {
		return
	}

	ai := c.makeDragonAttackInfo("Dragon of Dark Decay", dragonDarkDecayDmg, true)

	c.Core.QueueAttack(
		ai,
		combat.NewSingleTargetHit(c.Core.Combat.PrimaryTarget().Key()),
		0,
		0,
		c.a4DragonAttackCB,
	)

	c.Core.Tasks.Add(func() {
		c.dragonDarkDecayAttack(attackNum+1, src)
	}, dragonDarkDecayInterval)
}
