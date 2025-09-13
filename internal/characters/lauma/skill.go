package lauma

import (
	"math"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
	"github.com/genshinsim/gcsim/pkg/core/targets"
)

var skillFrames []int

const (
	skillInitHitmark     = 30
	skillInitHitmarkHold = 37
	skillTicks           = 8
	skillInterval        = 117
	skillFirstTickDelay  = 65
	skillKey             = "lauma-skill"
	particleICDKey       = "lauma-particle-icd"
)

func init() {
	skillFrames = frames.InitAbilSlice(41) // E -> D/J
}

// Ceil helper for tick timing
func ceil(x float64) int {
	return int(math.Ceil(x))
}

// Skill
// summoning a Frostgrove Sanctuary with different effects depending on whether you Tap or Hold.
func (c *char) Skill(p map[string]int) (action.Info, error) {
	skillPos := c.Core.Combat.Player()
	if p["hold"] == 1 && c.verdantDew > 0 {
		// Hold
		// Can be unleashed when you have at least 1 Verdant Dew.
		
		// Lauma consumes all Verdant Dew and intones a Hymn of Eternal Rest,
		// dealing one regular instance of AoE Dendro DMG and another instance of AoE Dendro DMG that is considered Lunar-Bloom DMG.
		// Each Verdant Dew consumed will give Lauma one stack of Moon Song.
		// Each time you Hold to cast an Elemental Skill, a maximum of 3 Verdant Dew can be consumed in this way.
		dewConsumed := min(c.verdantDew, 3)
		c.verdantDew -= dewConsumed
		c.moonSong += dewConsumed
		
		em := c.Stat(attributes.EM)
		ai1 := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Runo: Dawnless Rest of Karsikko (E/Hold/Hit 1)",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Dendro,
			Durability: 25,
			Mult:       skillHold1[c.TalentLvlSkill()],
		}
		
		c.Core.QueueAttack(
			ai1,
			combat.NewCircleHitOnTarget(skillPos, geometry.Point{Y: -1.5}, 5),
			skillInitHitmarkHold, skillInitHitmarkHold, c.particleCB,
		)
		// "considered Lunar-Bloom DMG area"
		ai2 := combat.AttackInfo{
			ActorIndex:       c.Index,
			Abil:             "Runo: Dawnless Rest of Karsikko (E/Hold/Hit 2)",
			AttackTag:        attacks.AttackTagLBDamage,
			ICDTag:           attacks.ICDTagNone,
			ICDGroup:         attacks.ICDGroupReactionB,
			StrikeType:       attacks.StrikeTypeDefault,
			Element:          attributes.Dendro,
			Durability:       0,
			IgnoreDefPercent: 1,
			FlatDmg:          (skillHold2[c.TalentLvlSkill()]*em*(1+c.LBBaseReactBonus(ai1)))*(1+((6*em)/(2000+em))+c.LBReactBonus(ai1)) + c.burstLBBuff,
		}
		snap := combat.Snapshot{
			CharLvl: c.Base.Level,
		}
		snap.Stats[attributes.CR] = c.Stat(attributes.CR)
		snap.Stats[attributes.CD] = c.Stat(attributes.CD)
		c.Core.QueueAttackWithSnap(
			ai2,
			snap,
			combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget().Pos(), nil, 6),
			skillInitHitmarkHold,
		)
		// "considered Lunar-Bloom DMG area end"
	} else {
		// Press E
		// dealing AoE Dendro DMG.
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Runo: Dawnless Rest of Karsikko (E/Press)",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Dendro,
			Durability: 25,
			Mult:       skillPress[c.TalentLvlSkill()],
		}
		
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(skillPos, geometry.Point{Y: -1.5}, 5),
			skillInitHitmark, skillInitHitmark, c.particleCB,
		)
	}

	// E duration and ticks are not affected by hitlag
	c.skillSrc = c.Core.F
	
	// C6: Reset sanctuary count when using skill
	if c.Base.Cons >= 6 {
		c.c6SanctuaryCount = 0
		// Remove C6 Pale Hymn stacks when using skill
		c.paleHymn -= c.c6PaleHymnCount
		c.c6PaleHymnCount = 0
	}
	
	for i := 0.0; i < skillTicks; i++ {
		c.Core.Tasks.Add(c.skillTick(c.skillSrc), skillFirstTickDelay+ceil(skillInterval*i))
	}
	c.AddStatus(skillKey, skillFirstTickDelay+ceil((skillTicks-1)*skillInterval), false)

	// Additionally, when Lauma's Elemental Skill or attacks from Frostgrove Sanctuary hit an opponent,
	// that opponent's Dendro RES and Hydro RES will be decreased for 10s.

	c.SetCD(action.ActionSkill, 12*60)
	c.a1() // Apply A1 moonsign buffs for 20s

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionSwap], // earliest cancel
		State:           action.SkillState,
	}, nil
}

// Particle generation callback for skill
func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	count := 1.0
	if c.Core.Rand.Float64() < 0.3 {
		count = 2.0
	}
	c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Dendro, c.ParticleDelay)
	c.AddStatus(particleICDKey, 3.3*60, true)
}

// Skill tick logic for DoT
func (c *char) skillTick(src int) func() {
	return func() {
		if src != c.skillSrc {
			return
		}
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Frostgrove Sanctuary Attack DMG (E/DoT)",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Dendro,
			Durability: 25,
			Mult:       skillDotATK[c.TalentLvlSkill()],
			FlatDmg:    skillDotEM[c.TalentLvlSkill()] * c.Stat(attributes.EM),
		}
		
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 1.5),
			0, 0, c.particleCB,
		)
		c.c4()
		c.c6SanctuaryBonus() // C6 additional damage on sanctuary attacks
	}
}
