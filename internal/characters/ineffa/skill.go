package ineffa

import (
	"math"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

var skillFrames []int

const (
	skillInitHitmark    = 25
	skillTicks          = 10
	skillInterval       = 117
	skillFirstTickDelay = 82
	skillKey            = "ineffa-skill"
	particleICDKey      = "ineffa-particle-icd"
)

func init() {
	skillFrames = frames.InitAbilSlice(41) // E -> D/J
}

// Ceil helper for tick timing
func ceil(x float64) int {
	return int(math.Ceil(x))
}

// Skill ability implementation
func (c *char) Skill(p map[string]int) (action.Info, error) {
	skillPos := c.Core.Combat.Player()
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Cleaning Mode: Carrier Frequency (E)",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       skill[c.TalentLvlSkill()],
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(skillPos, geometry.Point{Y: -1.5}, 5),
		skillInitHitmark, skillInitHitmark, c.particleCB,
	)

	// E duration and ticks are not affected by hitlag
	c.skillSrc = c.Core.F
	for i := 0.0; i < skillTicks; i++ {
		c.Core.Tasks.Add(c.skillTick(c.skillSrc), skillFirstTickDelay+ceil(skillInterval*i))
	}
	c.AddStatus(skillKey, skillFirstTickDelay+ceil((skillTicks-1)*skillInterval), false)

	c.genShield("ineffa-skill", c.shieldHP())
	c.SetCD(action.ActionSkill, 16*60)

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
	if c.Core.Rand.Float64() < 0.25 {
		return
	}
	c.Core.QueueParticle(c.Base.Key.String(), 1, attributes.Electro, c.ParticleDelay)
}

// Skill tick logic for DoT
func (c *char) skillTick(src int) func() {
	return func() {
		if src != c.skillSrc {
			return
		}
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Cleaning Mode: Carrier Frequency (E/DoT)",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Electro,
			Durability: 25,
			Mult:       skillDot[c.TalentLvlSkill()],
		}
		c.a1()
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 1.5),
			0, 0, c.particleCB,
		)
	}
}

