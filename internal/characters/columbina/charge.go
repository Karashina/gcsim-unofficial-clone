package columbina

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var chargeFrames []int

const chargeHitmark = 45

func init() {
	// Based on Mona (Hydro Catalyst) pattern with stub values
	chargeFrames = frames.InitAbilSlice(45) // CA -> D
	chargeFrames[action.ActionAttack] = 83  // CA -> N1
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	// Check if Verdant Dew is available for Moondew Cleanse
	if c.Core.Player.Verdant.Count() >= 1 {
		return c.moondewCleanse()
	}

	// Regular Charged Attack (Hydro DMG)
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Charge Attack",
		AttackTag:  attacks.AttackTagExtra,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
		Mult:       charge[c.TalentLvlAttack()],
	}

	// add windup if we're in idle or swap only
	windup := 14
	if c.Core.Player.CurrentState() == action.Idle || c.Core.Player.CurrentState() == action.SwapState {
		windup = 0
	}

	c.Core.QueueAttack(
		ai,
		combat.NewCircleHit(
			c.Core.Combat.Player(),
			c.Core.Combat.PrimaryTarget(),
			nil,
			3,
		),
		chargeHitmark-windup,
		chargeHitmark-windup,
	)

	return action.Info{
		Frames:          func(next action.Action) int { return chargeFrames[next] - windup },
		AnimationLength: chargeFrames[action.InvalidAction] - windup,
		CanQueueAfter:   chargeFrames[action.ActionSwap] - windup, // earliest cancel is before hitmark
		State:           action.ChargeAttackState,
	}, nil
}

// moondewCleanse performs the special charged attack when Verdant Dew is available
// Consumes 1 Verdant Dew to deal 3 instances of AoE Dendro DMG (considered as Lunar-Bloom DMG)
func (c *char) moondewCleanse() (action.Info, error) {
	// Check Moonridge Dew first, then Verdant Dew
	if c.moonridgeDew > 0 {
		c.moonridgeDew--
	} else {
		c.Core.Player.Verdant.Consume(1)
	}

	// 3 instances of Dendro DMG considered as Lunar-Bloom DMG
	for i := 0; i < 3; i++ {
		delay := chargeHitmark + i*8
		c.Core.Tasks.Add(func() {
			c.queueMoondewCleanseHit()
		}, delay)
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeHitmark,
		State:           action.ChargeAttackState,
	}, nil
}

// queueMoondewCleanseHit queues a single Moondew Cleanse hit (Lunar-Bloom DMG)
func (c *char) queueMoondewCleanseHit() {
	// Use AttackTagLBDamage for "considered as Lunar-Bloom DMG"
	ai := combat.AttackInfo{
		ActorIndex:       c.Index,
		Abil:             "Moondew Cleanse",
		AttackTag:        attacks.AttackTagLBDamage,
		ICDTag:           attacks.ICDTagNone,
		ICDGroup:         attacks.ICDGroupDefault,
		StrikeType:       attacks.StrikeTypeDefault,
		Element:          attributes.Dendro,
		Durability:       25,
		IgnoreDefPercent: 1,
	}

	// HP scaling with Lunar-Bloom formula
	em := c.Stat(attributes.EM)
	baseDmg := c.MaxHP() * moondewCleanse[c.TalentLvlAttack()] * (1 + c.LBBaseReactBonus(ai))
	emBonus := (6 * em) / (2000 + em)
	ai.FlatDmg = baseDmg * (1 + emBonus + c.LBReactBonus(ai)) * (1 + c.ElevationBonus(ai))

	snap := combat.Snapshot{
		CharLvl: c.Base.Level,
	}
	snap.Stats[attributes.CR] = c.Stat(attributes.CR)
	snap.Stats[attributes.CD] = c.Stat(attributes.CD)

	ap := combat.NewCircleHitOnTarget(
		c.Core.Combat.Player(),
		geometry.Point{Y: 1.5},
		4.0,
	)

	c.Core.QueueAttackWithSnap(ai, snap, ap, 0)

	// Emit Lunar-Bloom event for any hit
	enemies := c.Core.Combat.EnemiesWithinArea(ap, nil)
	if len(enemies) > 0 {
		ae := &combat.AttackEvent{Info: ai}
		c.Core.Events.Emit(event.OnLunarBloom, enemies[0], ae)
	}
}
