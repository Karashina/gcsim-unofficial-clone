package sethos

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/event"
)

var burstFrames []int

const (
	burstkey   = "sethos-burst"
	burstdrain = 5
	burstdelay = 2
)

func init() {
	burstFrames = frames.InitAbilSlice(57)
	burstFrames[action.ActionAttack] = 57
	burstFrames[action.ActionAim] = 57
	burstFrames[action.ActionSkill] = 57
	burstFrames[action.ActionDash] = 57
	burstFrames[action.ActionSwap] = 57
}

func (c *char) Burst(p map[string]int) (action.Info, error) {

	c.AddStatus(burstkey, 8*60, true) // activate for 8s

	c.ConsumeEnergy(5)
	c.SetCD(action.ActionBurst, 15*60)
	if c.Base.Cons >= 2 {
		c.AddStatus(c2burstkey, 600, true)
	}
	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap],
		State:           action.BurstState,
	}, nil
}

func (c *char) onExit() {
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(args ...interface{}) bool {
		prev := args[0].(int)
		next := args[1].(int)
		if prev == c.Index && next != c.Index {
			c.DeleteStatus(burstkey)
		}
		return false
	}, "sethos-exit")
}
