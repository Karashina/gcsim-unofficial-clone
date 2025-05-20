package escoffier

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
)

const (
	cookingMekStatus  = "cookingmek"
	cookingMekHitmark = 5

	cookingMekTickInterval = 59
)

func (c *char) spawnCookingMek() {
	c.cookingMekSrc = c.Core.F
	c.AddStatus(cookingMekStatus, 20*60, true)
	c.QueueCharTask(c.cookingMekAttack(c.cookingMekSrc), cookingMekTickInterval)
}

func (c *char) cookingMekAttack(src int) func() {
	return func() {
		if c.cookingMekSrc != src {
			return
		}
		if !c.StatusIsActive(cookingMekStatus) {
			return
		}
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Frosty Parfait (E-DoT)",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagElementalArt,
			ICDGroup:   attacks.ICDGroupEscoffierCookingMek,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Cryo,
			Durability: 25,
			Mult:       skillCookingMek[c.TalentLvlSkill()],
		}
		ap := combat.NewCircleHit(c.Core.Combat.Player().Pos(), c.Core.Combat.PrimaryTarget(), nil, 1)
		c.Core.QueueAttack(ai, ap, cookingMekHitmark, cookingMekHitmark)
		c.QueueCharTask(c.cookingMekAttack(src), cookingMekTickInterval)
	}
}
