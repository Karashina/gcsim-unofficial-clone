package flins

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/core/reactions"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	c4IcdKey = "flins-c4-icd"
	c6IcdKey = "flins-c6-icd"
)

// C1: Adds reaction bonus mod for all characters
func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}
	for _, char := range c.Core.Player.Chars() {
		amt := min(0.5, c.TotalAtk()/100*0.025)
		char.AddLCReactBonusMod(character.LCReactBonusMod{
			Base: modifier.NewBase("flins-c1", 20*60),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				return amt, false
			},
		})
	}
}

// C2: Generates shield and triggers dummy attack
func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}
	c.genShield("flins-skill", c.shieldHP())
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Flins C2 Dummy",
		FlatDmg:    0,
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 99), 0, 0)
}

// C4: Adds energy on LC damage, with ICD
func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		a := args[1].(*combat.AttackEvent)
		if a.Info.AttackTag != attacks.AttackTagLCDamage {
			return false
		}
		if c.StatusIsActive(c4IcdKey) {
			return false
		}
		c.AddEnergy("Flins C4", 5)
		c.AddStatus(c4IcdKey, 4*60, true)
		return false
	}, "flins-c4")
}

// C6: Triggers dummy attack if C1 buff is active and LC triggers, with ICD
func (c *char) c6() {
	if c.Base.Cons < 6 {
		return
	}
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		a := args[1].(*combat.AttackEvent)
		if a.Info.AttackTag != attacks.AttackTagLCDamage {
			return false
		}
		if a.Info.Abil != string(reactions.LunarCharged) {
			return false
		}
		if !c.ReactBonusModIsActive("flins-c1") {
			return false
		}
		if c.StatusIsActive(c6IcdKey) {
			return false
		}
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Flins C6 Dummy",
			FlatDmg:    0,
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 99), 0, 0)
		c.AddStatus(c6IcdKey, 3.5*60, true)
		return false
	}, "flins-c6")
}
