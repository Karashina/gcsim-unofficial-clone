package mavuika

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/geometry"
)

var bikeDashFrames []int
var bikeDashHitmark = 6

func init() {
	bikeDashFrames = frames.InitAbilSlice(20) // dash
}

func (c *char) Dash(p map[string]int) (action.Info, error) {
	if c.nightsoulState.HasBlessing() {
		c.reduceNightsoulPoints(10)

		ai := combat.AttackInfo{
			ActorIndex:     c.Index,
			Abil:           "Flamestrider Sprint DMG",
			AttackTag:      attacks.AttackTagNone,
			AdditionalTags: []attacks.AdditionalTag{attacks.AdditionalTagNightsoul},
			ICDTag:         attacks.ICDTagNormalAttack,
			ICDGroup:       attacks.ICDGroupDefault,
			StrikeType:     attacks.StrikeTypeBlunt,
			Element:        attributes.Pyro,
			Durability:     25,
			Mult:           bikeplunge[c.TalentLvlSkill()],
			IgnoreInfusion: true,
		}
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1}, 2),
			bikeDashHitmark,
			bikeDashHitmark,
		)

		// assuming doesn't contribute to dash CD
		return action.Info{
			Frames:          frames.NewAbilFunc(bikeDashFrames),
			AnimationLength: bikeDashFrames[action.InvalidAction],
			CanQueueAfter:   bikeDashFrames[action.ActionSwap],
			State:           action.DashState,
		}, nil
	}
	return c.Character.Dash(p)
}
