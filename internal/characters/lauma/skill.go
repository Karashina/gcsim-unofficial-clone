package lauma

import (
	"math"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
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
		c.verdantDew = 0
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
			skillInitHitmarkHold, skillInitHitmarkHold, c.particleCB, c.shredCB,
		)
		// "considered Lunar-Bloom DMG area"
		ai2 := combat.AttackInfo{
			ActorIndex:       c.Index,
			Abil:             "Runo: Dawnless Rest of Karsikko (E/Hold/Hit 2)",
			AttackTag:        attacks.AttackTagLBDamage,
			StrikeType:       attacks.StrikeTypeDefault,
			Element:          attributes.Dendro,
			IgnoreDefPercent: 1,
			FlatDmg:          (skillHold2[c.TalentLvlSkill()]*em*(1+c.LBBaseReactBonus(ai1)))*(1+((6*em)/(2000+em))+c.LBReactBonus(ai1)) + c.burstLBBuff*c.c6mult,
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
			c.shredCB,
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
			skillInitHitmark, skillInitHitmark, c.particleCB, c.shredCB,
		)
	}

	// E duration and ticks are not affected by hitlag
	c.skillSrc = c.Core.F

	for i := 0.0; i < skillTicks; i++ {
		c.Core.Tasks.Add(c.skillTick(c.skillSrc), skillFirstTickDelay+ceil(skillInterval*i))
	}
	c.AddStatus(skillKey, skillFirstTickDelay+ceil((skillTicks-1)*skillInterval), false)

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
			combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 6.5),
			0, 0, c.particleCB, c.shredCB,
		)
		c.c4()

		if c.Base.Cons >= 6 {
			// Deal additional AoE Dendro DMG equal to 185% of EM
			em := c.Stat(attributes.EM)
			c6ai := combat.AttackInfo{
				ActorIndex:       c.Index,
				Abil:             "Frostgrove Sanctuary (C6 Bonus)",
				AttackTag:        attacks.AttackTagLBDamage,
				StrikeType:       attacks.StrikeTypeDefault,
				Element:          attributes.Dendro,
				IgnoreDefPercent: 1,
			}

			snapc6 := combat.Snapshot{
				CharLvl: c.Base.Level,
			}
			c6ai.FlatDmg = (1.85*em*(1+c.LBBaseReactBonus(c6ai)))*(1+((6*em)/(2000+em))+c.LBReactBonus(c6ai)) + c.burstLBBuff*c.c6mult // 185% of EM
			snapc6.Stats[attributes.CR] = c.Stat(attributes.CR)
			snapc6.Stats[attributes.CD] = c.Stat(attributes.CD)

			c.Core.QueueAttackWithSnap(
				c6ai,
				snapc6,
				combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget().Pos(), nil, 6.5),
				1,
				c.shredCB,
			)

			c.paleHymn += 3 // Gain 2+1 Pale Hymn stacks
			c.AddStatus("pale-hymn-window", 15*60, true)
		}
	}
}

func (c *char) shredCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	e, ok := a.Target.(*enemy.Enemy)
	if !ok {
		return
	}
	e.AddResistMod(combat.ResistMod{
		Base:  modifier.NewBaseWithHitlag("lauma-e-dendro", 10*60),
		Ele:   attributes.Dendro,
		Value: -1 * skillRESShred[c.TalentLvlSkill()],
	})
}

