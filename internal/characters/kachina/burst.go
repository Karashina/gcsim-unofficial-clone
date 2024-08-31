package kachina

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

var burstFramesNormal []int

const (
	burstKey = "kachina-burst"
)

func init() {
	burstFramesNormal = frames.InitAbilSlice(67)
}

const burstHitmark = 40

func (c *char) Burst(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Time to Get Serious!",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		PoiseDMG:   150,
		Element:    attributes.Geo,
		Durability: 25,
		Mult:       burst[c.TalentLvlBurst()],
		UseDef:     true,
	}

	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 10),
		burstHitmark, burstHitmark)

	c.AddStatus(burstKey, 12*60, true)
	c.SetCD(action.ActionBurst, 18*60)
	c.BurstBuffField()
	c.c2()
	c.ConsumeEnergy(6)

	return action.Info{
		Frames:          func(next action.Action) int { return burstFramesNormal[next] },
		AnimationLength: burstFramesNormal[action.InvalidAction],
		CanQueueAfter:   burstFramesNormal[action.ActionAttack],
		State:           action.BurstState,
	}, nil
}

func (c *char) BurstBuffField() func() {
	return func() {
		area := combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 10)
		if c.StatusIsActive(burstKey) && !c.Core.Combat.Player().IsWithinArea(area) {
			return
		}

		if c.Base.Cons >= 4 {
			buff := make([]float64, attributes.EndStatType)

			switch min(4, c.Core.Combat.EnemyCount()) {
			case 1:
				buff[attributes.DEFP] = 0.08
			case 2:
				buff[attributes.DEFP] = 0.12
			case 3:
				buff[attributes.DEFP] = 0.16
			case 4:
				buff[attributes.DEFP] = 0.20
			default:
				buff[attributes.DEFP] = 0.00
			}

			active := c.Core.Player.ActiveChar()
			active.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag("kachina-c4", 120),
				AffectedStat: attributes.DEFP,
				Amount: func() ([]float64, bool) {
					return buff, true
				},
			})
		}
		// tick every 0.3s
		c.Core.Tasks.Add(c.BurstBuffField(), 18)
	}
}
