package pyro

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/targets"
)

var (
	skillPressFrames [][]int
	skillHoldFrames  [][]int

	skillPressHitmark = 21
	skillHoldHitmark  = 53
)

const (
	skillPressCdStart               = 19
	skillPressDoTInterval           = 66
	skillPressNightsoulconsumestart = 35

	skillHoldCdStart               = 56
	skillHoldDoTHitmark            = 12
	skillHoldNightsoulconsumestart = 61

	skillHoldKey    = "travelerpyro-skill-hold"
	skillHoldIcdKey = "travelerpyro-skill-hold-icd"

	particleICDKey = "travelerpyro-particle-icd"
)

func init() {
	// Tap E
	skillPressFrames = make([][]int, 2)

	// Male
	skillPressFrames[0] = skillPressFrames[1] // Tap E -> E

	// Female
	skillPressFrames[1] = frames.InitAbilSlice(36) // Tap E -> E/Q/Walk

	// Hold E
	skillHoldFrames = make([][]int, 2)

	// Male
	skillHoldFrames[0] = skillHoldFrames[1] // Short Hold E -> Swap

	// Female
	skillHoldFrames[1] = frames.InitAbilSlice(60) // Short Hold E -> Swap
}

func (c *Traveler) reduceNightsoulPoints(val float64) {
	c.nightsoulState.ConsumePoints(val)

	if c.nightsoulState.Points() <= 0.00001 {
		c.cancelNightsoul()
	}
}

func (c *Traveler) cancelNightsoul() {
	c.nightsoulState.ExitBlessing()
	c.DeleteStatus(skillHoldKey)
	c.nightsoulSrc = -1
}

func (c *Traveler) nightsoulPointReduceFunc(src int) func() {
	return func() {
		if c.nightsoulSrc != src {
			return
		}

		if !c.nightsoulState.HasBlessing() {
			return
		}

		c.reduceNightsoulPoints(0.75)

		// reduce 0.75 point per 6f
		c.QueueCharTask(c.nightsoulPointReduceFunc(src), 6)
	}
}

func (c *Traveler) skillPress(hitmark, cdStart int, skillFrames [][]int) (action.Info, error) {
	if c.nightsoulState.HasBlessing() {
		c.cancelNightsoul()
	}
	c.nightsoulState.ClearPoints()
	c.nightsoulState.EnterBlessing(42)
	c.QueueCharTask(c.blazingThreshold(c.nightsoulSrc), hitmark)
	c.QueueCharTask(c.nightsoulPointReduceFunc(c.nightsoulSrc), skillPressNightsoulconsumestart)

	c.c2Count = 0
	c.AddStatus(c2Key, 12*60, true)
	c.SetCDWithDelay(action.ActionSkill, 18*60, cdStart)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames[c.gender]),
		AnimationLength: skillFrames[c.gender][action.InvalidAction],
		CanQueueAfter:   skillFrames[c.gender][action.ActionDash], // earliest cancel
		State:           action.SkillState,
	}, nil
}

func (c *Traveler) blazingThreshold(src int) func() {
	return func() {

		if c.nightsoulSrc != src {
			return
		}

		if !c.nightsoulState.HasBlessing() {
			return
		}

		ai := combat.AttackInfo{
			ActorIndex:     c.Index,
			Abil:           "Blazing Threshold",
			AttackTag:      attacks.AttackTagElementalArt,
			AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
			ICDTag:         attacks.ICDTagElementalArt,
			ICDGroup:       attacks.ICDGroupDefault,
			StrikeType:     attacks.StrikeTypeDefault,
			Element:        attributes.Pyro,
			Durability:     25,
			Mult:           skillPress[c.TalentLvlSkill()],
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player().Pos(), nil, 5), 0, 0, c.skillParticleCB)
		c.QueueCharTask(c.blazingThreshold(src), skillPressDoTInterval)
	}
}

func (c *Traveler) skillHold() (action.Info, error) {
	if c.nightsoulState.HasBlessing() {
		c.cancelNightsoul()
	}
	c.nightsoulState.ClearPoints()
	c.nightsoulState.EnterBlessing(42)
	c.QueueCharTask(c.nightsoulPointReduceFunc(c.nightsoulSrc), skillHoldNightsoulconsumestart)

	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Flowfire Blade: Hold DMG",
		AttackTag:      attacks.AttackTagElementalArt,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagElementalArtPyro,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Pyro,
		Durability:     25,
		Mult:           skillhold[c.TalentLvlSkill()],
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player().Pos(), nil, 5), skillHoldHitmark, skillHoldHitmark)

	c.AddStatus(skillHoldKey, -1, false)
	c.c2Count = 0
	c.AddStatus(c2Key, 12*60, true)
	c.SetCDWithDelay(action.ActionSkill, 18*60, skillHoldCdStart)

	return action.Info{
		Frames:          func(next action.Action) int { return skillHoldFrames[c.gender][next] },
		AnimationLength: skillHoldFrames[c.gender][action.InvalidAction],
		CanQueueAfter:   skillHoldFrames[c.gender][action.ActionJump], // earliest cancel
		State:           action.SkillState,
	}, nil
}

func (c *Traveler) scorchingThreshold() {
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)

		if ae.Info.ActorIndex != c.Core.Player.ActiveChar().Index {
			return false
		}

		if !c.nightsoulState.HasBlessing() {
			return false
		}

		if !c.StatusIsActive(skillHoldKey) {
			return false
		}

		if c.StatusIsActive(skillHoldIcdKey) {
			return false
		}

		ai := combat.AttackInfo{
			ActorIndex:     c.Index,
			Abil:           "Scorching Threshold",
			AttackTag:      attacks.AttackTagElementalArt,
			AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
			ICDTag:         attacks.ICDTagElementalArt,
			ICDGroup:       attacks.ICDGroupDefault,
			StrikeType:     attacks.StrikeTypeDefault,
			Element:        attributes.Pyro,
			Durability:     25,
			Mult:           skillholddot[c.TalentLvlSkill()],
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player().Pos(), nil, 5), skillHoldDoTHitmark, skillHoldDoTHitmark, c.skillParticleCB)
		c.AddStatus(skillHoldIcdKey, 3*60, false)
		return false
	}, "traveller-skill-hold")
}

func (c *Traveler) Skill(p map[string]int) (action.Info, error) {
	hold := p["hold"]
	if hold >= 1 {
		return c.skillHold()
	} else {
		return c.skillPress(
			skillPressHitmark,
			skillPressCdStart,
			skillPressFrames,
		)
	}
}

func (c *Traveler) skillParticleCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, 3*60, true)

	c.Core.QueueParticle(c.Base.Key.String(), 1, attributes.Pyro, c.ParticleDelay)
}

func (c *Traveler) durWatcher() {
	c.Core.Events.Subscribe(event.OnTick, func(args ...interface{}) bool {
		if c.nightsoulState.HasBlessing() {
			if c.Core.F-c.nightsoulSrc >= 12*60 {
				c.cancelNightsoul()
			}
		}
		return false
	}, "traveller-skill-duration-check")
}
