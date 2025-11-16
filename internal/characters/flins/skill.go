package flins

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
	skillFrames     []int
	northlandFrames []int
)

const (
	northlandHitmark = 23
	skillKey         = "flins-skill"
	northlandKey     = "flins-northland"
	northlandCdKey   = "flins-northland-cd"
	particleICDKey   = "flins-particle-icd"
)

func init() {
	skillFrames = frames.InitAbilSlice(25)     // E -> D/J
	northlandFrames = frames.InitAbilSlice(37) // E -> D/J
}

// E
/*
Flins enters Manifest Flame form and obtains skillKey. This form has the following characteristics:
· Flins's Normal and Charged Attack's element is converted into Electro and he is unable to use Plunging Attacks.
· His Elemental Skill: "Ancient Rite: Arcane Light" is replaced with the special Elemental Skill: "Northland Spearstorm".

Northland Spearstorm
Deals AoE Electro DMG and gives him northlandKey for the next 6s.

---ADDITIONAL INFO FOR COPILOT---
for implementation of "Flins's Normal and Charged Attack's element is converted into Electro and he is unable to use Plunging Attacks",
this should be handled in attack.go and charge.go. for additional example, see implementation of cyno.
*/
func (c *char) Skill(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(skillKey) {
		aiNorthland := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Northland Spearstorm DMG (E)",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Electro,
			Durability: 25,
			Mult:       skill[c.TalentLvlSkill()],
		}
		c.Core.QueueAttack(
			aiNorthland,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: -1.5}, 5),
			northlandHitmark, northlandHitmark, c.particleCB,
		)

		c.AddStatus(northlandKey, 6*60, true)
		c.AddStatus(northlandCdKey, c.northlandCD, true)

		// C2: Set status for additional damage on next Normal Attack
		if c.Base.Cons >= 2 {
			c.AddStatus(c2NorthlandKey, 6*60, true)
		}

		return action.Info{
			Frames:          frames.NewAbilFunc(northlandFrames),
			AnimationLength: northlandFrames[action.InvalidAction],
			CanQueueAfter:   northlandFrames[action.ActionSwap], // earliest cancel
			State:           action.SkillState,
		}, nil
	} else {
		c.AddStatus(skillKey, 10*60, true) // 10s
		c.SetCD(action.ActionSkill, 16*60)

		return action.Info{
			Frames:          frames.NewAbilFunc(skillFrames),
			AnimationLength: skillFrames[action.InvalidAction],
			CanQueueAfter:   skillFrames[action.ActionSwap], // earliest cancel
			State:           action.SkillState,
		}, nil
	}
}

// Particle generation callback for skill - called every time an Electro-converted attack is landed
func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, 2.1*60, true)
	c.Core.QueueParticle(c.Base.Key.String(), 1, attributes.Electro, c.ParticleDelay)
}

