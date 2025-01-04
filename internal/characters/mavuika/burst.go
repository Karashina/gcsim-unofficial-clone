package mavuika

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
)

const (
	burstHitmark = 116 // adjusted to swap frame
	spiritIcdKey = "mavuika-spirit-icd"
	BurstKey     = "mavuika-burst"
	BurstCDKey   = "mavuika-burst-cd"
)

var (
	burstFrames []int
)

func init() {
	burstFrames = frames.InitAbilSlice(125)
}

func (c *char) Burst(p map[string]int) (action.Info, error) {

	c.consumedspirit = 0
	c.consumedspirit = c.fightingspirit

	c2buff := 0.0
	if c.Base.Cons >= 2 {
		c2buff = 1.2
	}

	ai := combat.AttackInfo{
		ActorIndex:     c.Index,
		Abil:           "Hour of Burning Skies",
		AttackTag:      attacks.AttackTagElementalBurst,
		AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
		ICDTag:         attacks.ICDTagNone,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeDefault,
		Element:        attributes.Pyro,
		Durability:     25,
		Mult:           burst[c.TalentLvlBurst()] + burstbonus[c.TalentLvlBurst()]*c.consumedspirit + c2buff,
	}
	burstArea := combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 7)

	c.Core.QueueAttack(ai, burstArea, burstHitmark, burstHitmark)

	c.QueueCharTask(func() { c.AddStatus(BurstKey, 10*60, true) }, 110)

	c.enterNightsoul(10, 94)

	c.AddStatus(bikeKey, -1, false)
	c.c2DefModRemove()
	c.c6DefModAdd()
	c.c6()

	c.a4()
	c.AddStatus(BurstCDKey, 18*60, false)
	c.QueueCharTask(func() { c.fightingspirit = 0 }, 2)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // earliest cancel
		State:           action.BurstState,
	}, nil
}

func (c *char) GainFightingSpirit() {
	mult := 1.00
	if c.Base.Cons >= 1 {
		mult = 1.25
	}
	c.Core.Events.Subscribe(event.OnNightsoulConsume, func(args ...interface{}) bool {
		amt := args[1].(float64)

		c.fightingspirit += float64(amt * mult)
		if c.fightingspirit >= c.maxfightingspirit {
			c.fightingspirit = c.maxfightingspirit
		}
		c.c1()

		return false
	}, "mavuika-fightingspirit-nightsoul")
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)

		if c.StatusIsActive(spiritIcdKey) {
			return false
		}
		if ae.Info.AttackTag != attacks.AttackTagNormal {
			return false
		}

		c.fightingspirit += float64(1.5 * mult)
		if c.fightingspirit >= c.maxfightingspirit {
			c.fightingspirit = c.maxfightingspirit
		}

		c.AddStatus(spiritIcdKey, 0.1*60, false)
		c.c1()

		return false
	}, "mavuika-fightingspirit-attack")
}
