package mavuika

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	c.Core.Events.Subscribe(event.OnNightsoulBurst, func(args ...interface{}) bool {
		m := make([]float64, attributes.EndStatType)
		m[attributes.ATKP] = 0.3
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("mavuika-a1", 10*60),
			AffectedStat: attributes.ATKP,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
		return false
	}, "mavuika-a1-nightsoul")
}

func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	started := c.Core.F
	maxamt := min(0.4, 0.002*c.consumedspirit)
	for _, char := range c.Core.Player.Chars() {
		this := char
		this.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("mavuika-a4", 20*60),
			Amount: func(_ *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
				// char must be active
				if c.Core.Player.Active() != this.Index {
					return nil, false
				}
				// floor time elapsed
				dmg := maxamt - maxamt*(float64((c.Core.F-started)/60)/20)
				if c.Base.Cons >= 4 {
					dmg = maxamt + 0.1
				}
				c.a4buff[attributes.DmgP] = dmg
				return c.a4buff, true
			},
		})
	}
}
