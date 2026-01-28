package ineffa

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

var burstFrames []int

const (
	hitmark = 119
)

func init() {
	burstFrames = frames.InitAbilSlice(114)
}

// Burst ability implementation
func (c *char) Burst(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Supreme Instruction: Cyclonic Exterminator (Q)",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       burst[c.TalentLvlBurst()],
	}
	c.QueueCharTask(func() {
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: -1.5}, 7),
			0, 0,
		)
	}, hitmark)

	c.a4()
	c.c2()
	c.SetCD(action.ActionBurst, 15*60)

	// resummon E
	c.skillSrc = c.Core.F
	for i := 0.0; i < skillTicks; i++ {
		c.Core.Tasks.Add(c.skillTick(c.skillSrc), skillFirstTickDelay+ceil(skillInterval*i))
	}
	c.AddStatus(skillKey, skillFirstTickDelay+ceil((skillTicks-1)*skillInterval), false)

	c.ConsumeEnergy(4)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // earliest cancel
		State:           action.BurstState,
	}, nil
}
