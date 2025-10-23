package nefer

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
)

var chargeFrames []int

const (
	chargeHitmark = 22
)

func init() {
	chargeFrames = frames.InitAbilSlice(55)
	chargeFrames[action.ActionDash] = chargeHitmark
	chargeFrames[action.ActionJump] = chargeHitmark
	chargeFrames[action.ActionWalk] = 53
	chargeFrames[action.ActionSkill] = 1
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	// Charged attack
	/*
		Nefer enters the Slither state, consuming 1.82 Stamina per 6 frames to move rapidly forward for up to 2.5s.
		When the hold duration ends, or Stamina runs out, Nefer will exit the Slither state and deal Dendro DMG to opponents this attack consumes remaining stamina up to 50.
		When in the Shadow Dance state, max attack Stamina consumption is decreased to 25.
	*/
	// Elemental Skill interaction
	/*
		When nefer is on the Shadow Dance state and party have at least 1 Verdant Dew, Nefer's Charged Attacks will be replaced with the special Charged Attack "Phantasm Performance", which will not consume Stamina.
		Nefer will deal "Phantasm Performance n-Hit DMG (Nefer)" 2 times and "Phantasm Performance n-Hit DMG (Shades)" 3 times. DMG dealt by the shades is considered Lunar-Bloom DMG.
		after Phantasm Performance 1-Hit DMG (Shades) is used, 1 Verdant Dew will be consumed.
		TO COPILOT: actual DMG will be implemented in lunarbloomhook.go so just queue dummy attack like ineffa/lauma/flins.
	*/
	// A4 interaction
	/*
	   When Nefer is in the Shadow Dance state and any party member triggers a Lunar-Bloom reaction, Nefer's Slither state will provide additional Verdant Dew for 5s.
	   Every points of Nefer's Elemental Mastery beyond 500 will strengthen this additional provision effect by 0.1%. The maximum increase that can be achieved this way is 50%.
	*/
	if p["hold"] > 0 {
		// hold
		dur := p["hold"]
		chargeFrames[action.ActionSkill] = 1 // using the Elemental Skill while Nefer is in the Slither state will not cause her to exit the state.
		c.Core.Player.Verdant.StartCharge(dur)
		c.Core.Tasks.Add(func() { c.Core.Player.Verdant.GetGainBonus() }, dur)
		c.Core.Player.Verdant.SetGainBonus(min(0.001*float64(c.Stat(attributes.EM)), 0.5))
	}

	return action.Info{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeHitmark,
		State:           action.ChargeAttackState,
	}, nil
}
