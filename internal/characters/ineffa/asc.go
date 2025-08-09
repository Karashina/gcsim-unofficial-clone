package ineffa

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

// A0: Adds base reaction bonus mod for all characters
func (c *char) a0() {
	for _, char := range c.Core.Player.Chars() {
		char.AddStatus("LC-Key", -1, false)
		char.AddLCBaseReactBonusMod(character.LCBaseReactBonusMod{
			Base: modifier.NewBase("Assemblage Hub (A0)", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				maxval := 0.14
				return min(maxval, c.TotalAtk()/100*0.007), true
			},
		})
	}
}

// A1: Triggers dummy attack if ascension 1 and skill is active
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	if !c.StatusIsActive(skillKey) {
		return
	}
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Ineffa A1 Dummy",
		FlatDmg:    0,
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 99), 0, 0)
}

// A4: Adds EM stat mod for all characters based on total ATK
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	for _, char := range c.Core.Player.Chars() {
		m := make([]float64, attributes.EndStatType)
		m[attributes.EM] = c.TotalAtk() * 0.06
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("ineffa-a4", 20*60),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				if char.Index != c.Index {
					if c.Core.Player.Active() != char.Index {
						return nil, false
					}
				}
				return m, true
			},
		})
	}
}
