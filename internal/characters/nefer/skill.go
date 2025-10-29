package nefer

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
	"github.com/genshinsim/gcsim/pkg/core/targets"
	"github.com/genshinsim/gcsim/pkg/reactable"
)

var (
	skillFrames []int
)

const (
	skillHitmark = 26
	skillKey     = "nefer-skill"
)

func init() {
	skillFrames = frames.InitAbilSlice(31)
}

// Elemental Skill: AoE Dendro + Shadow Dance status; grants charges. In Shadow Dance, Verdant Dew enables Phantasm Performance.
func (c *char) Skill(p map[string]int) (action.Info, error) {

	// Skill DMG has both ATK and EM scaling
	c.QueueCharTask(func() {
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Skill Initial DMG (E)",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagElementalArt,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Dendro,
			Durability: 25,
			FlatDmg:    skillatk[c.TalentLvlSkill()]*c.Stat(attributes.ATK) + skillem[c.TalentLvlSkill()]*c.Stat(attributes.EM),
		}

		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0}, 5),
			0,
			0,
			c.particleCB,
		)

		// Enter Shadow Dance state
		c.AddStatus(skillKey, 10*60, true) // 10s duration

		// If Moonsign Ascendant: convert existing Dendro Cores to Seeds of Deceit and set 15s conversion window
		if c.MoonsignAscendant {
			// set status window
			c.AddStatus("nefer-seed-convert", 15*60, true)
			// convert existing dendro cores
			for _, g := range c.Core.Combat.Gadgets() {
				if g == nil {
					continue
				}
				if g.GadgetTyp() == combat.GadgetTypDendroCore {
					// type assert to reactable.DendroCore and mark as seed
					if dc, ok := g.(*reactable.DendroCore); ok {
						dc.IsSeed = true
						// disable explosions and reaction triggers
						dc.Gadget.OnExpiry = nil
						dc.Gadget.OnKill = nil
					}
				}
			}
		}

		// C2: Gain 2 stacks of Veil of Falsehood when using Elemental Skill
		if c.Base.Cons >= 2 && c.Base.Ascension >= 1 {
			maxStacks := 5.0
			if c.a1count < maxStacks-1 {
				c.a1count += 2
			} else if c.a1count < maxStacks {
				c.a1count = maxStacks
			}
			c.AddStatus("veil-of-falsehood", 14*60, true) // 9s base + 5s from C2
		}
	}, skillHitmark)

	c.SetCDWithDelay(action.ActionSkill, 9*60, skillHitmark)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionSwap],
		State:           action.SkillState,
	}, nil
}

func (c *char) particleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	count := 2.0
	if c.Core.Rand.Float64() < 0.667 {
		count = 3
	}
	c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Dendro, c.ParticleDelay)
}
