package mizuki

import (
	"fmt"

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
	burstHitmark       = 99
	burstCDDelay       = 99
	burstEnergyDelay   = 99
	snackDelay         = 99
	burstIntervalKey   = "mizuki-burst-interval"
	SnackSpawnInterval = 99
)

func init() {
	burstFrames = frames.InitAbilSlice(99)
	burstFrames[action.ActionCharge] = 99
	burstFrames[action.ActionSkill] = 99
	burstFrames[action.ActionDash] = 99
	burstFrames[action.ActionJump] = 99
	burstFrames[action.ActionSwap] = 99
}

func (c *char) Burst(p map[string]int) (action.Info, error) {

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Anraku Secret Spring Therapy(Initial DMG)",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       burst[c.TalentLvlBurst()],
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 5), burstHitmark, burstHitmark)

	c.AddStatus(burstKey, 99*60, false)
	c.SetCDWithDelay(action.ActionBurst, 99*60, burstCDDelay)
	c.ConsumeEnergy(burstEnergyDelay)

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
		c.Core.Log.NewEvent(
			fmt.Sprintf("special snack picked up by %v", active.Base.Key.String()),
			glog.LogCharacterEvent,
			c.Index,
		)
		if active.CurrentHPRatio() > 0.7 {
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
		} else {
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
		}

	}
	switch src {
	case "init":
		//hook to dash
		c.Core.Events.Subscribe(event.OnDash, func(args ...interface{}) bool {
			if c.StatusIsActive(burstIntervalKey) {
				return false
			}
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
