package lauma

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
)

var burstFrames []int

const (
	buffapply = 104
)

func init() {
	burstFrames = frames.InitAbilSlice(116)
}

// Burst
// gain 18 stacks of Pale Hymn.
// Additionally, if Lauma uses her Elemental Burst while she has Moon Song,or she gains Moon Song within 15s of using her Elemental Burst,
// she will consume all Moon Song stacks and gain 6 stacks of Pale Hymn for every Moon Song stack consumed.
// This effect can only be triggered once for each Elemental Burst used, including the 15 seconds following its use.
// Pale Hymn
// When nearby party members deal Bloom, Hyperbloom, Burgeon, or Lunar-Bloom DMG,
// 1 stack of Pale Hymn will be consumed and the DMG dealt will be increased based on Lauma's Elemental Mastery.
// If this DMG hits multiple opponents at once, then multiple stacks of Pale Hymn will be consumed, depending on how many opponents are hit.
// The duration for each stack of Pale Hymn is counted independently.
// if Lauma is C2 or higher,
// Pale Hymn effects are increased: All nearby party members' Bloom, Hyperbloom, and Burgeon DMG is further increased by 500% of Lauma's Elemental Mastery,
// and their Lunar-Bloom DMG is further increased by 400% of Lauma's Elemental Mastery.
func (c *char) Burst(p map[string]int) (action.Info, error) {

	c.SetCD(action.ActionBurst, 15*60)
	c.ConsumeEnergy(7)
	c.c2()

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // earliest cancel
		State:           action.BurstState,
	}, nil
}
