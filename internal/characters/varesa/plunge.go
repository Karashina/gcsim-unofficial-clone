package varesa

import (
	"errors"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/player"
	"github.com/genshinsim/gcsim/pkg/core/targets"
)

var highPlungeFrames []int
var lowPlungeFrames []int
var highPlungeHitmark = 47

const lowPlungeHitmark = 45
const collisionHitmark = lowPlungeHitmark - 6
const lowPlungeRadius = 3.0
const highPlungeRadius = 3.5

func init() {
	// low_plunge -> x
	lowPlungeFrames = frames.InitAbilSlice(65)
	lowPlungeFrames[action.ActionAttack] = 57
	lowPlungeFrames[action.ActionCharge] = 57 - 11
	lowPlungeFrames[action.ActionSkill] = 57
	lowPlungeFrames[action.ActionBurst] = 57
	lowPlungeFrames[action.ActionDash] = lowPlungeHitmark
	lowPlungeFrames[action.ActionWalk] = 64
	lowPlungeFrames[action.ActionSwap] = 48

	// high_plunge -> x
	highPlungeFrames = frames.InitAbilSlice(67)
	highPlungeFrames[action.ActionAttack] = 58
	highPlungeFrames[action.ActionCharge] = 58 - 9
	highPlungeFrames[action.ActionSkill] = 59
	highPlungeFrames[action.ActionBurst] = 59
	highPlungeFrames[action.ActionDash] = highPlungeHitmark
	lowPlungeFrames[action.ActionWalk] = 66
	highPlungeFrames[action.ActionSwap] = 50
}

func (c *char) setValesaFrames() {
	highPlungeHitmark = 35
	highPlungeFrames = frames.InitAbilSlice(46)
	highPlungeFrames[action.ActionAttack] = 36
	highPlungeFrames[action.ActionDash] = highPlungeHitmark
}

// Low Plunge attack damage queue generator
// Use the "collision" optional argument if you want to do a falling hit on the way down
// Default = 0
func (c *char) LowPlungeAttack(p map[string]int) (action.Info, error) {
	defer c.Core.Player.SetAirborne(player.Grounded)
	switch c.Core.Player.Airborne() {
	case player.AirborneVaresa:
		c.setValesaFrames()
		if c.StatusIsActive(fieryPassionKey) {
			return c.highPlungeFP(p), nil
		} else {
			return c.highPlunge(p), nil
		}
	case player.AirborneXianyun:
		if c.StatusIsActive(fieryPassionKey) {
			return c.lowPlungeFP(p), nil
		} else {
			return c.lowPlunge(p), nil
		}
	default:
		return action.Info{}, errors.New("low_plunge can only be used while airborne")
	}
}

func (c *char) lowPlunge(p map[string]int) action.Info {
	collision, ok := p["collision"]
	if !ok {
		collision = 0 // Whether or not collision hit
	}

	if collision > 0 {
		c.plungeCollision(collisionHitmark)
	}

	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Low Plunge",
		AttackTag:      attacks.AttackTagPlunge,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagNone,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Electro,
		Durability:     25,
		Mult:           lowPlunge[c.TalentLvlAttack()],
	}
	c.c4Plunge()
	ai.Mult += c.a1PlungeBuff()
	ai.FlatDmg += c.c4buff
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, lowPlungeRadius),
		lowPlungeHitmark,
		lowPlungeHitmark,
		c.plungeCB,
		c.c2CB(),
		c.a1Cancel,
	)

	return action.Info{
		Frames:          frames.NewAbilFunc(lowPlungeFrames),
		AnimationLength: lowPlungeFrames[action.InvalidAction],
		CanQueueAfter:   lowPlungeFrames[action.ActionDash],
		State:           action.PlungeAttackState,
	}
}

// High Plunge attack damage queue generator
// Use the "collision" optional argument if you want to do a falling hit on the way down
// Default = 0
func (c *char) HighPlungeAttack(p map[string]int) (action.Info, error) {
	defer c.Core.Player.SetAirborne(player.Grounded)
	switch c.Core.Player.Airborne() {
	case player.AirborneVaresa:
		c.setValesaFrames()
		if c.StatusIsActive(fieryPassionKey) {
			return c.highPlungeFP(p), nil
		} else {
			return c.highPlunge(p), nil
		}
	case player.AirborneXianyun:
		if c.StatusIsActive(fieryPassionKey) {
			return c.highPlungeFP(p), nil
		} else {
			return c.highPlunge(p), nil
		}
	default:
		return action.Info{}, errors.New("high_plunge can only be used while airborne")
	}
}

func (c *char) highPlunge(p map[string]int) action.Info {
	collision, ok := p["collision"]
	if !ok {
		collision = 0 // Whether or not collision hit
	}

	if collision > 0 {
		c.plungeCollision(collisionHitmark)
	}

	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "High Plunge",
		AttackTag:      attacks.AttackTagPlunge,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagNone,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Electro,
		Durability:     25,
		Mult:           highPlunge[c.TalentLvlAttack()],
	}
	c.c4Plunge()
	ai.Mult += c.a1PlungeBuff()
	ai.FlatDmg += c.c4buff
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, highPlungeRadius),
		highPlungeHitmark,
		highPlungeHitmark,
		c.plungeCB,
		c.c2CB(),
		c.a1Cancel,
	)

	return action.Info{
		Frames:          frames.NewAbilFunc(highPlungeFrames),
		AnimationLength: highPlungeFrames[action.InvalidAction],
		CanQueueAfter:   highPlungeFrames[action.ActionDash],
		State:           action.PlungeAttackState,
	}
}

// Plunge normal falling attack damage queue generator
// Standard - Always part of high/low plunge attacks
func (c *char) plungeCollision(delay int) {
	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Plunge Collision",
		AttackTag:      attacks.AttackTagPlunge,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagNone,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Electro,
		Durability:     0,
		Mult:           collision[c.TalentLvlAttack()],
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 1.5), delay, delay)
}

//---------------------Fiery Passion--------------------------//

func (c *char) lowPlungeFP(p map[string]int) action.Info {
	collision, ok := p["collision"]
	if !ok {
		collision = 0 // Whether or not collision hit
	}

	if collision > 0 {
		c.plungeCollision(collisionHitmark)
	}

	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Low Plunge (Fiery Passion)",
		AttackTag:      attacks.AttackTagPlunge,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagNone,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Electro,
		Durability:     25,
		Mult:           lowPlungefp[c.TalentLvlAttack()],
	}
	c.c4Plunge()
	ai.Mult += c.a1PlungeBuff()
	ai.FlatDmg += c.c4buff
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, lowPlungeRadius),
		lowPlungeHitmark,
		lowPlungeHitmark,
		c.plungeCB,
		c.c2CB(),
		c.a1Cancel,
	)
	c.nightsoulState.ConsumePoints(40)
	c.QueueCharTask(func() {
		c.nightsoulState.ExitBlessing()
		c.DeleteStatus(fieryPassionKey)
	}, 3*60)

	return action.Info{
		Frames:          frames.NewAbilFunc(lowPlungeFrames),
		AnimationLength: lowPlungeFrames[action.InvalidAction],
		CanQueueAfter:   lowPlungeFrames[action.ActionDash],
		State:           action.PlungeAttackState,
	}
}

func (c *char) highPlungeFP(p map[string]int) action.Info {
	collision, ok := p["collision"]
	if !ok {
		collision = 0 // Whether or not collision hit
	}

	if collision > 0 {
		c.plungeCollision(collisionHitmark)
	}

	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "High Plunge (Fiery Passion)",
		AttackTag:      attacks.AttackTagPlunge,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagNone,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Electro,
		Durability:     25,
		Mult:           highPlungefp[c.TalentLvlAttack()],
	}
	c.c4Plunge()
	ai.Mult += c.a1PlungeBuff()
	ai.FlatDmg += c.c4buff
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, highPlungeRadius),
		highPlungeHitmark,
		highPlungeHitmark,
		c.plungeCB,
		c.c2CB(),
		c.a1Cancel,
	)
	c.nightsoulState.ConsumePoints(40)
	c.QueueCharTask(func() {
		c.nightsoulState.ExitBlessing()
		c.DeleteStatus(fieryPassionKey)
	}, 3*60)

	return action.Info{
		Frames:          frames.NewAbilFunc(highPlungeFrames),
		AnimationLength: highPlungeFrames[action.InvalidAction],
		CanQueueAfter:   highPlungeFrames[action.ActionDash],
		State:           action.PlungeAttackState,
	}
}

func (c *char) plungeCB(a combat.AttackCB) {
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if !c.StatusIsActive(fieryPassionKey) {
		c.nightsoulState.GeneratePoints(25)
	}
}
