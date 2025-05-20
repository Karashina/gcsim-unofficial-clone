package escoffier

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	a1Key         = "Rehab Diet (A1)"
	a1InitialHeal = 154
	a1Interval    = 60
)

func (c *char) a1(src int) func() {
	return func() {
		if c.Base.Ascension < 1 {
			return
		}
		if c.a1Src != src {
			return
		}
		if !c.StatusIsActive(a1Key) {
			return
		}
		c4mult := 1.0
		if c.c4() {
			c4mult = 2.0
		}
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index,
			Target:  c.Core.Player.ActiveChar().Index,
			Message: "Better to Salivate Than Medicate (A1)",
			Src:     c.TotalAtk() * 1.3824 * c4mult,
			Bonus:   c.Stat(attributes.Heal),
		})

		c.QueueCharTask(c.a1(src), a1Interval)
	}
}

func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}

	chars := c.Core.Player.Chars()
	count := make(map[attributes.Element]int)
	for _, this := range chars {
		count[this.Base.Element]++
	}

	shred := 0.00
	switch count[attributes.Cryo] + count[attributes.Hydro] {
	case 1:
		shred = -0.05
	case 2:
		shred = -0.10
	case 3:
		shred = -0.15
	case 4:
		shred = -0.55
		c.c1()
	default:
		shred = -0.00
	}

	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != c.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagElementalArt && atk.Info.AttackTag != attacks.AttackTagElementalBurst {
			return false
		}

		t, ok := args[0].(*enemy.Enemy)
		if !ok {
			return false
		}
		t.AddResistMod(combat.ResistMod{
			Base:  modifier.NewBaseWithHitlag("escoffier-a4-cryo", 12*60),
			Ele:   attributes.Cryo,
			Value: shred,
		})
		t.AddResistMod(combat.ResistMod{
			Base:  modifier.NewBaseWithHitlag("escoffier-a4-hydro", 12*60),
			Ele:   attributes.Hydro,
			Value: shred,
		})

		return false
	}, "escoffier-a4")
}
