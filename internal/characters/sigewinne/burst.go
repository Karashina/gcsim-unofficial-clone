package sigewinne

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
)

var burstFrames []int

const (
	burstkey     = "sigewinne-burst"
	bursthitmark = 101
	burstdelay   = 2
)

func init() {
	burstFrames = frames.InitAbilSlice(230)
	burstFrames[action.ActionAttack] = 230
	burstFrames[action.ActionAim] = 230
	burstFrames[action.ActionSkill] = 230
	burstFrames[action.ActionDash] = 95
	burstFrames[action.ActionJump] = 95
	burstFrames[action.ActionSwap] = 230
}

func (c *char) Burst(p map[string]int) (action.Info, error) {

	c.burstHitSrc = 0
	c.burstCounter = 0
	if c.Base.Cons < 4 {
		c.AddStatus(burstkey, 230, true)
	} else {
		c.AddStatus(burstkey, 410, true)
		burstFrames = frames.InitAbilSlice(410)
		burstFrames[action.ActionAttack] = 410
		burstFrames[action.ActionAim] = 410
		burstFrames[action.ActionSkill] = 410
		burstFrames[action.ActionDash] = 95
		burstFrames[action.ActionJump] = 95
		burstFrames[action.ActionSwap] = 410
	}
	c.QueueCharTask(c.burstTick(c.burstHitSrc), 95)

	c.c2()
	c.SetCDWithDelay(action.ActionBurst, 18*60, 1)
	c.ConsumeEnergy(5)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionJump], // earliest cancel
		State:           action.BurstState,
		OnRemoved:       func(next action.AnimationState) { c.c2Remove() },
	}, nil
}

func (c *char) burstTick(src int) func() {
	return func() {
		if c.burstHitSrc != src {
			return
		}
		if c.Core.Player.Active() != c.Index {
			return
		}
		if !c.StatusIsActive(burstkey) {
			return
		}
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Super Saturated Syringing",
			AttackTag:  attacks.AttackTagElementalBurst,
			ICDTag:     attacks.ICDTagExtraAttack,
			ICDGroup:   attacks.ICDGroupSigewinneBurst,
			StrikeType: attacks.StrikeTypePierce,
			Element:    attributes.Hydro,
			Durability: 25,
			FlatDmg:    burst[c.TalentLvlBurst()] * c.MaxHP(),
		}
		c.Core.QueueAttack(
			ai,
			combat.NewBoxHitOnTarget(c.Core.Combat.Player(), nil, 5, 15),
			1,
			1,
			c.c2Resist,
		)
		c.burstCounter++
		c.burstHitSrc++
		c.QueueCharTask(c.burstTick(c.burstHitSrc), 25)
	}
}
