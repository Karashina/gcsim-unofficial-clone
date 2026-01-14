package dhalia

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

const (
	burstkey = "dhalia-burst"
	a1IcdKey = "dhalia-a1-icd"
)

var (
	burstFrames []int
)

func init() {
	burstFrames = frames.InitAbilSlice(54) // Q -> Walk
	burstFrames[action.ActionDash] = 54
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Secret Art: Surprise Dispatch",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Dendro,
		Durability: 50,
		Mult:       burst[c.TalentLvlBurst()],
	}
	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 0.5}, 2.6)
	c.Core.QueueAttack(
		ai,
		ap,
		30,
		30,
	)

	c.genShield("dhalia-burst", c.shieldHP())
	c.nacount = 0
	c.favoniusfavor = 0
	c.AddStatus(burstkey, c.burstdur, true)
	c.a4()
	c.c6()
	c.SetCD(action.ActionBurst, 15*60)
	c.ConsumeEnergy(5)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash],
		State:           action.BurstState,
	}, nil
}

func (c *char) generateFavonianFavor() {
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		if !c.StatusIsActive(burstkey) {
			return false
		}
		atk := args[1].(*combat.AttackEvent)
		switch atk.Info.AttackTag {
		case attacks.AttackTagNormal:
			c.nacount++
			if c.nacount >= 4 {
				c.favoniusfavor++
				if c.favoniusfavor >= 4 {
					c.favoniusfavor = 4
				}
			}
			c.c1()
		default:
			return false
		}
		return false
	}, "dhalia-ff-na")
	if c.Base.Ascension >= 1 {
		c.Core.Events.Subscribe(event.OnFrozen, func(args ...interface{}) bool {
			if !c.StatusIsActive(burstkey) {
				return false
			}
			if c.StatusIsActive(a1IcdKey) {
				return false
			}
			c.AddStatus(a1IcdKey, 8*60, true)
			c.favoniusfavor += 2
			if c.favoniusfavor >= 4 {
				c.favoniusfavor = 4
			}
			c.c1()
			return false
		}, "dhalia-ff-frozen")
	}
}

