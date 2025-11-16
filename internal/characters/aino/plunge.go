package aino

import (
	"errors"

	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var highPlungeFrames []int
var lowPlungeFrames []int

const lowPlungeHitmark = 44
const highPlungeHitmark = 43
const collisionHitmark = 39

const lowPlungePoiseDMG = 100.0
const lowPlungeRadius = 3.0

const highPlungePoiseDMG = 150.0
const highPlungeRadius = 5.0

func init() {
	lowPlungeFrames = frames.InitAbilSlice(68)
	lowPlungeFrames[action.ActionAttack] = 60
	lowPlungeFrames[action.ActionSkill] = 59
	lowPlungeFrames[action.ActionBurst] = 60
	lowPlungeFrames[action.ActionDash] = lowPlungeHitmark
	lowPlungeFrames[action.ActionSwap] = 58

	highPlungeFrames = frames.InitAbilSlice(67)
	highPlungeFrames[action.ActionAttack] = 60
	highPlungeFrames[action.ActionSkill] = 59
	highPlungeFrames[action.ActionBurst] = 59
	highPlungeFrames[action.ActionDash] = highPlungeHitmark
	highPlungeFrames[action.ActionSwap] = 58
}

func (c *char) LowPlungeAttack(p map[string]int) (action.Info, error) {
	defer c.Core.Player.SetAirborne(c.Core.Player.Airborne() - 1)
	switch c.Core.Player.Airborne() {
	case 0:
		return action.Info{}, errors.New("low plunge can only be used while airborne")
	case 1:
		// collision
		if _, ok := p["collision"]; ok {
			c.plungeCollision(collisionHitmark)
		}
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Low Plunge",
			AttackTag:  attacks.AttackTagPlunge,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeBlunt,
			Element:    attributes.Physical,
			Durability: 25,
			Mult:       lowPlunge[c.TalentLvlAttack()],
			PoiseDMG:   lowPlungePoiseDMG,
		}
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, lowPlungeRadius),
			lowPlungeHitmark,
			lowPlungeHitmark,
		)

		return action.Info{
			Frames:          frames.NewAbilFunc(lowPlungeFrames),
			AnimationLength: lowPlungeFrames[action.InvalidAction],
			CanQueueAfter:   lowPlungeFrames[action.ActionDash],
			State:           action.PlungeAttackState,
		}, nil
	default:
		return action.Info{}, errors.New("plunge must be executed while airborne and within 1 meter of the ground")
	}
}

func (c *char) HighPlungeAttack(p map[string]int) (action.Info, error) {
	defer c.Core.Player.SetAirborne(c.Core.Player.Airborne() - 1)
	switch c.Core.Player.Airborne() {
	case 0:
		return action.Info{}, errors.New("high plunge can only be used while airborne")
	case 1:
		return action.Info{}, errors.New("high plunge must be executed while airborne and above 1 meter of the ground")
	default:
		// collision
		if _, ok := p["collision"]; ok {
			c.plungeCollision(collisionHitmark)
		}
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "High Plunge",
			AttackTag:  attacks.AttackTagPlunge,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeBlunt,
			Element:    attributes.Physical,
			Durability: 25,
			Mult:       highPlunge[c.TalentLvlAttack()],
			PoiseDMG:   highPlungePoiseDMG,
		}
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, highPlungeRadius),
			highPlungeHitmark,
			highPlungeHitmark,
		)

		return action.Info{
			Frames:          frames.NewAbilFunc(highPlungeFrames),
			AnimationLength: highPlungeFrames[action.InvalidAction],
			CanQueueAfter:   highPlungeFrames[action.ActionDash],
			State:           action.PlungeAttackState,
		}, nil
	}
}

func (c *char) plungeCollision(delay int) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Plunge Collision",
		AttackTag:  attacks.AttackTagPlunge,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeSlash,
		Element:    attributes.Physical,
		Durability: 25,
		Mult:       collision[c.TalentLvlAttack()],
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, 1), delay, delay)
}

