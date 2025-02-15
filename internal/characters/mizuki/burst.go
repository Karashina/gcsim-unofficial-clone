package mizuki

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

var burstFrames []int

const (
	burstKey           = "mizuki-burst"
	burstHitmark       = 94
	burstCDDelay       = 95
	burstEnergyDelay   = 5
	snackDelay         = 190
	burstIntervalKey   = "mizuki-burst-interval"
	SnackSpawnInterval = 60
)

func init() {
	burstFrames = frames.InitAbilSlice(103)
	burstFrames[action.ActionCharge] = 103
	burstFrames[action.ActionSkill] = 103
	burstFrames[action.ActionDash] = 103
	burstFrames[action.ActionJump] = 103
	burstFrames[action.ActionSwap] = 103
}

func (c *char) Burst(p map[string]int) (action.Info, error) {

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Anraku Secret Spring Therapy(Initial DMG)",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 50,
		Mult:       burst[c.TalentLvlBurst()],
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 5), burstHitmark, burstHitmark)

	c.AddStatus(burstKey, 12*60, false)
	c.SetCDWithDelay(action.ActionBurst, 15*60, burstCDDelay)
	c.ConsumeEnergy(burstEnergyDelay)
	c.c4Count = 0

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash], // earliest cancel
		State:           action.BurstState,
	}, nil
}

func (c *char) snackHandler(src string) {
	snackfunc := func() {
		active := c.Core.Player.ActiveChar()
		if active.CurrentHPRatio() > 0.7 || c.Base.Cons >= 4 {
			// snacc attacc
			c.Core.Tasks.Add(func() {
				ai := combat.AttackInfo{
					ActorIndex: c.Index,
					Abil:       "Anraku Secret Spring Therapy(Munen Shockwave)",
					AttackTag:  attacks.AttackTagElementalBurst,
					ICDTag:     attacks.ICDTagElementalBurst,
					ICDGroup:   attacks.ICDGroupDefault,
					StrikeType: attacks.StrikeTypeDefault,
					Element:    attributes.Anemo,
					Durability: 25,
					Mult:       burstSnack[c.TalentLvlBurst()],
				}
				c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 5), 0, 0)
			}, snackDelay)
		}
		if active.CurrentHPRatio() <= 0.7 || c.Base.Cons >= 4 {
			// snacc heal
			c.Core.Tasks.Add(func() {
				heal := burstHealC[c.TalentLvlBurst()] + burstHeal[c.TalentLvlBurst()]*c.Stat(attributes.EM)
				c.Core.Player.Heal(info.HealInfo{
					Caller:  c.Index,
					Target:  c.Core.Player.Active(),
					Message: "Anraku Secret Spring Therapy(Heal)",
					Src:     heal,
					Bonus:   c.Stat(attributes.Heal),
				})
			}, snackDelay)
			c.c4()
		}
	}
	switch src {
	case "init":
		//hook to dash
		c.Core.Events.Subscribe(event.OnDash, func(args ...interface{}) bool {
			if c.StatusIsActive(burstIntervalKey) {
				return false
			}
			c.Core.Log.NewEvent(
				"special snack picked up by dash",
				glog.LogCharacterEvent,
				c.Index,
			)
			c.AddStatus(burstIntervalKey, SnackSpawnInterval, false)
			snackfunc()
			return false
		}, "mizuki-burst-snack-ondash")
	case "skill":
		if !c.StatusIsActive(burstKey) {
			return
		}
		c.Core.Log.NewEvent(
			"special snack picked up by mizuki E",
			glog.LogCharacterEvent,
			c.Index,
		)
		snackfunc()
	default:
		if !c.StatusIsActive(burstKey) {
			return
		}
		if c.StatusIsActive(burstIntervalKey) {
			return
		}
		c.AddStatus(burstIntervalKey, SnackSpawnInterval, false)
		c.Core.Log.NewEvent(
			"special snack picked up by undefined move",
			glog.LogCharacterEvent,
			c.Index,
		)
		snackfunc()
	}
}
