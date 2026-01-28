package columbina

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player"
)

var (
	plungeFrames     []int
	highPlungeFrames []int
	lowPlungeFrames  []int
)

// TODO: Frame data not measured, using stub values
const (
	collisionHitmark  = 999
	highPlungeHitmark = 999
	lowPlungeHitmark  = 999
)

func init() {
	// TODO: Frame data not measured, using stub values
	plungeFrames = frames.InitAbilSlice(999)
	plungeFrames[action.ActionAttack] = 999
	plungeFrames[action.ActionSkill] = 999
	plungeFrames[action.ActionBurst] = 999
	plungeFrames[action.ActionDash] = 999
	plungeFrames[action.ActionJump] = 999
	plungeFrames[action.ActionSwap] = 999

	highPlungeFrames = frames.InitAbilSlice(999)
	highPlungeFrames[action.ActionAttack] = 999
	highPlungeFrames[action.ActionSkill] = 999
	highPlungeFrames[action.ActionBurst] = 999
	highPlungeFrames[action.ActionDash] = 999
	highPlungeFrames[action.ActionJump] = 999
	highPlungeFrames[action.ActionSwap] = 999

	lowPlungeFrames = frames.InitAbilSlice(999)
	lowPlungeFrames[action.ActionAttack] = 999
	lowPlungeFrames[action.ActionSkill] = 999
	lowPlungeFrames[action.ActionBurst] = 999
	lowPlungeFrames[action.ActionDash] = 999
	lowPlungeFrames[action.ActionJump] = 999
	lowPlungeFrames[action.ActionSwap] = 999
}

func (c *char) HighPlungeAttack(p map[string]int) (action.Info, error) {
	defer c.Core.Player.SetAirborne(player.Grounded)

	// Collision
	_, ok := p["collision"]
	if !ok {
		aiCollision := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Plunge Collision (P)",
			AttackTag:  attacks.AttackTagPlunge,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeSlash,
			Element:    attributes.Hydro,
			Durability: 0,
			Mult:       collision[c.TalentLvlAttack()],
		}
		c.Core.QueueAttack(
			aiCollision,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 1.0),
			collisionHitmark,
			collisionHitmark,
		)
	}

	// High Plunge
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "High Plunge (HP)",
		AttackTag:  attacks.AttackTagPlunge,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		Element:    attributes.Hydro,
		Durability: 25,
		Mult:       highPlunge[c.TalentLvlAttack()],
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 3.5),
		highPlungeHitmark,
		highPlungeHitmark,
	)

	return action.Info{
		Frames:          frames.NewAbilFunc(highPlungeFrames),
		AnimationLength: highPlungeFrames[action.InvalidAction],
		CanQueueAfter:   highPlungeHitmark,
		State:           action.PlungeAttackState,
	}, nil
}

func (c *char) LowPlungeAttack(p map[string]int) (action.Info, error) {
	defer c.Core.Player.SetAirborne(player.Grounded)

	// Collision
	_, ok := p["collision"]
	if !ok {
		aiCollision := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Plunge Collision (P)",
			AttackTag:  attacks.AttackTagPlunge,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeSlash,
			Element:    attributes.Hydro,
			Durability: 0,
			Mult:       collision[c.TalentLvlAttack()],
		}
		c.Core.QueueAttack(
			aiCollision,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 1.0),
			collisionHitmark,
			collisionHitmark,
		)
	}

	// Low Plunge
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Low Plunge (LP)",
		AttackTag:  attacks.AttackTagPlunge,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		Element:    attributes.Hydro,
		Durability: 25,
		Mult:       lowPlunge[c.TalentLvlAttack()],
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 3.0),
		lowPlungeHitmark,
		lowPlungeHitmark,
	)

	return action.Info{
		Frames:          frames.NewAbilFunc(lowPlungeFrames),
		AnimationLength: lowPlungeFrames[action.InvalidAction],
		CanQueueAfter:   lowPlungeHitmark,
		State:           action.PlungeAttackState,
	}, nil
}
