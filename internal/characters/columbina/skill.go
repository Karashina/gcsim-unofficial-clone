package columbina

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

// TODO: Frame data not measured, using stub values
const (
	skillHitmark          = 999
	gravityRippleDuration = 25 * 60 // 25 seconds
	gravityRippleInterval = 4 * 60  // 4 seconds
	gravityLimit          = 60
	gravityPerTick        = 20
	gravityTickInterval   = 2 * 60 // 2 seconds
)

func init() {
	skillFrames = frames.InitAbilSlice(999)
	skillFrames[action.ActionAttack] = 999
	skillFrames[action.ActionCharge] = 999
	skillFrames[action.ActionSkill] = 999
	skillFrames[action.ActionBurst] = 999
	skillFrames[action.ActionDash] = 999
	skillFrames[action.ActionJump] = 999
	skillFrames[action.ActionSwap] = 999
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	// Initial skill damage (AoE Hydro DMG)
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Eternal Tides",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
	}
	ai.FlatDmg = c.MaxHP() * skillDmg[c.TalentLvlSkill()]

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5)
	c.Core.QueueAttack(ai, ap, skillHitmark, skillHitmark, c.particleCB)

	// Set Gravity Ripple active
	c.AddStatus(skillKey, gravityRippleDuration, true)
	c.AddStatus(gravityRippleKey, gravityRippleDuration, true)

	// Reset gravity
	c.gravity = 0
	c.gravityLC = 0
	c.gravityLB = 0
	c.gravityLCrs = 0

	// Start Gravity Ripple ticks
	c.gravityRippleSrc = c.Core.F
	c.gravityRippleExp = c.Core.F + gravityRippleDuration
	c.Core.Tasks.Add(c.gravityRippleTick(c.Core.F), gravityRippleInterval)

	// C1: Trigger Gravity Interference effect on skill cast (once per 15s)
	if c.Base.Cons >= 1 {
		c.c1OnSkill()
	}

	// Set cooldown (17s)
	c.SetCDWithDelay(action.ActionSkill, 17*60, skillHitmark)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash],
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

	// Average 1.333 particles (0/1/2 with weighted distribution)
	count := 1.0
	if c.Core.Rand.Float64() < 0.333 {
		count = 2.0
	}
	c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Hydro, c.ParticleDelay)
}

// gravityRippleTick handles periodic Gravity Ripple damage
func (c *char) gravityRippleTick(src int) func() {
	return func() {
		if c.gravityRippleSrc != src {
			return
		}
		if !c.StatusIsActive(gravityRippleKey) {
			return
		}

		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Gravity Ripple",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Hydro,
			Durability: 25,
		}
		ai.FlatDmg = c.MaxHP() * gravityRippleDmg[c.TalentLvlSkill()]

		ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 6)
		c.Core.QueueAttack(ai, ap, 0, 0, c.particleCB)

		// Schedule next tick
		c.Core.Tasks.Add(c.gravityRippleTick(src), gravityRippleInterval)
	}
}

// subscribeToLunarReactions subscribes to Lunar reaction events for Gravity accumulation
func (c *char) subscribeToLunarReactions() {
	// Subscribe to Lunar-Charged
	c.Core.Events.Subscribe(event.OnLunarCharged, func(args ...interface{}) bool {
		if !c.StatusIsActive(gravityRippleKey) {
			return false
		}
		c.accumulateGravity("lc")
		return false
	}, "columbina-gravity-lc")

	// Subscribe to Lunar-Bloom
	c.Core.Events.Subscribe(event.OnLunarBloom, func(args ...interface{}) bool {
		if !c.StatusIsActive(gravityRippleKey) {
			return false
		}
		c.accumulateGravity("lb")
		return false
	}, "columbina-gravity-lb")

	// Subscribe to Lunar-Crystallize
	c.Core.Events.Subscribe(event.OnLunarCrystallize, func(args ...interface{}) bool {
		if !c.StatusIsActive(gravityRippleKey) {
			return false
		}
		c.accumulateGravity("lcrs")
		return false
	}, "columbina-gravity-lcrs")
}

// accumulateGravity accumulates Gravity from Lunar reactions
func (c *char) accumulateGravity(lunarType string) {
	// Check if New Moon's Omen ICD allows accumulation (20 Gravity per 2s)
	if c.StatusIsActive(newMoonOmenKey) {
		return
	}

	// C2: Rate of accumulating Gravity increases by 34%
	gravityGain := gravityPerTick
	if c.Base.Cons >= 2 {
		gravityGain = int(float64(gravityGain) * 1.34)
	}

	c.gravity += gravityGain
	switch lunarType {
	case "lc":
		c.gravityLC += gravityGain
	case "lb":
		c.gravityLB += gravityGain
	case "lcrs":
		c.gravityLCrs += gravityGain
	}

	// Cap at gravity limit
	if c.gravity > gravityLimit {
		c.gravity = gravityLimit
	}

	c.AddStatus(newMoonOmenKey, gravityTickInterval, false)

	c.Core.Log.NewEvent("gravity accumulated", glog.LogCharacterEvent, c.Index).
		Write("type", lunarType).
		Write("gravity", c.gravity).
		Write("lc", c.gravityLC).
		Write("lb", c.gravityLB).
		Write("lcrs", c.gravityLCrs)

	// Check if Gravity is at limit
	if c.gravity >= gravityLimit {
		c.triggerGravityInterference()
	}
}

// triggerGravityInterference triggers Gravity Interference based on dominant Lunar type
func (c *char) triggerGravityInterference() {
	dominantType := c.getDominantLunarType()

	c.Core.Log.NewEvent("gravity interference triggered", glog.LogCharacterEvent, c.Index).
		Write("dominant_type", dominantType).
		Write("gravity", c.gravity)

	// A1: Gain Lunacy effect
	if c.Base.Ascension >= 1 {
		c.a1OnGravityInterference()
	}

	// C2: Gain Lunar Brilliance
	if c.Base.Cons >= 2 {
		c.c2OnGravityInterference(dominantType)
	}

	// C4: Restore energy and DMG bonus
	if c.Base.Cons >= 4 {
		c.c4OnGravityInterference(dominantType)
	}

	switch dominantType {
	case "lc":
		c.gravityInterferenceLC()
	case "lb":
		c.gravityInterferenceLB()
	case "lcrs":
		c.gravityInterferenceLCrs()
	}

	// Reset gravity
	c.gravity = 0
	c.gravityLC = 0
	c.gravityLB = 0
	c.gravityLCrs = 0
}

// gravityInterferenceLC deals Electro AoE DMG (considered Lunar-Charged DMG)
func (c *char) gravityInterferenceLC() {
	ai := combat.AttackInfo{
		ActorIndex:       c.Index,
		Abil:             "Gravity Interference (Lunar-Charged)",
		AttackTag:        attacks.AttackTagLCDamage,
		ICDTag:           attacks.ICDTagNone,
		ICDGroup:         attacks.ICDGroupDefault,
		StrikeType:       attacks.StrikeTypeDefault,
		Element:          attributes.Electro,
		Durability:       0,
		IgnoreDefPercent: 1,
	}

	// HP scaling with Lunar-Charged formula
	em := c.Stat(attributes.EM)
	baseDmg := c.MaxHP() * gravityInterfLC[c.TalentLvlSkill()] * (1 + c.LCBaseReactBonus(ai))
	emBonus := (6 * em) / (2000 + em)
	ai.FlatDmg = baseDmg * (1 + emBonus + c.LCReactBonus(ai)) * (1 + c.ElevationBonus(ai))

	// C4 bonus: LC=12.5%, LB=2.5%, LCrs=12.5% of Max HP
	if c.Base.Cons >= 4 && c.c4ICD <= c.Core.F {
		switch c.c4DominantType {
		case "lc":
			ai.FlatDmg += c.MaxHP() * 0.125
		case "lb":
			ai.FlatDmg += c.MaxHP() * 0.025
		case "lcrs":
			ai.FlatDmg += c.MaxHP() * 0.125
		}
	}

	snap := combat.Snapshot{CharLvl: c.Base.Level}
	snap.Stats[attributes.CR] = c.Stat(attributes.CR)
	snap.Stats[attributes.CD] = c.Stat(attributes.CD)

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 8)
	c.Core.QueueAttackWithSnap(ai, snap, ap, 10)

	// Emit event
	enemies := c.Core.Combat.EnemiesWithinArea(ap, nil)
	if len(enemies) > 0 {
		ae := &combat.AttackEvent{Info: ai}
		c.Core.Events.Emit(event.OnLunarCharged, enemies[0], ae)
	}
}

// gravityInterferenceLB fires 5 Moondew Sigils dealing Dendro DMG (considered Lunar-Bloom DMG)
func (c *char) gravityInterferenceLB() {
	for i := 0; i < 5; i++ {
		delay := 10 + i*6
		c.Core.Tasks.Add(func() {
			ai := combat.AttackInfo{
				ActorIndex:       c.Index,
				Abil:             "Gravity Interference (Lunar-Bloom)",
				AttackTag:        attacks.AttackTagLBDamage,
				ICDTag:           attacks.ICDTagNone,
				ICDGroup:         attacks.ICDGroupDefault,
				StrikeType:       attacks.StrikeTypeDefault,
				Element:          attributes.Dendro,
				Durability:       0,
				IgnoreDefPercent: 1,
			}

			// HP scaling with Lunar-Bloom formula
			em := c.Stat(attributes.EM)
			baseDmg := c.MaxHP() * gravityInterfLB[c.TalentLvlSkill()] * (1 + c.LBBaseReactBonus(ai))
			emBonus := (6 * em) / (2000 + em)
			ai.FlatDmg = baseDmg * (1 + emBonus + c.LBReactBonus(ai)) * (1 + c.ElevationBonus(ai))

			// C4 bonus: 2.5% of Max HP applies to each hit
			if c.Base.Cons >= 4 && c.c4ICD <= c.Core.F {
				ai.FlatDmg += c.MaxHP() * 0.025
			}

			snap := combat.Snapshot{CharLvl: c.Base.Level}
			snap.Stats[attributes.CR] = c.Stat(attributes.CR)
			snap.Stats[attributes.CD] = c.Stat(attributes.CD)

			closest := c.Core.Combat.ClosestEnemy(c.Core.Combat.Player().Pos())
			if closest == nil {
				return
			}
			ap := combat.NewCircleHitOnTarget(closest.Pos(), nil, 2)
			c.Core.QueueAttackWithSnap(ai, snap, ap, 0)

			// Emit event
			ae := &combat.AttackEvent{Info: ai}
			c.Core.Events.Emit(event.OnLunarBloom, closest, ae)
		}, delay)
	}
}

// gravityInterferenceLCrs deals Geo AoE DMG (considered Lunar-Crystallize DMG)
func (c *char) gravityInterferenceLCrs() {
	ai := combat.AttackInfo{
		ActorIndex:       c.Index,
		Abil:             "Gravity Interference (Lunar-Crystallize)",
		AttackTag:        attacks.AttackTagLCrsDamage,
		ICDTag:           attacks.ICDTagNone,
		ICDGroup:         attacks.ICDGroupDefault,
		StrikeType:       attacks.StrikeTypeDefault,
		Element:          attributes.Geo,
		Durability:       0,
		IgnoreDefPercent: 1,
	}

	// HP scaling with Lunar-Crystallize formula
	em := c.Stat(attributes.EM)
	baseDmg := c.MaxHP() * gravityInterfLCrs[c.TalentLvlSkill()] * (1 + c.LCrsBaseReactBonus(ai))
	emBonus := (6 * em) / (2000 + em)
	ai.FlatDmg = baseDmg * (1 + emBonus + c.LCrsReactBonus(ai)) * (1 + c.ElevationBonus(ai))

	// C4 bonus
	if c.Base.Cons >= 4 && c.c4ICD <= c.Core.F {
		ai.FlatDmg += c.MaxHP() * 0.125
	}

	snap := combat.Snapshot{CharLvl: c.Base.Level}
	snap.Stats[attributes.CR] = c.Stat(attributes.CR)
	snap.Stats[attributes.CD] = c.Stat(attributes.CD)

	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 8)
	c.Core.QueueAttackWithSnap(ai, snap, ap, 10)

	// Emit event
	enemies := c.Core.Combat.EnemiesWithinArea(ap, nil)
	if len(enemies) > 0 {
		ae := &combat.AttackEvent{Info: ai}
		c.Core.Events.Emit(event.OnLunarCrystallize, enemies[0], ae)
	}
}
