package citlali

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	a1Duration = 12 * 60
	a1IcdKey   = "citlali-a1-icd"
)

func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	c.Core.Events.Subscribe(event.OnMelt, func(args ...interface{}) bool {
		t, ok := args[0].(*enemy.Enemy)
		if !ok {
			return false
		}
		c.a1Handle(t)
		return false
	}, "citlali-a1-melt")
	c.Core.Events.Subscribe(event.OnFrozen, func(args ...interface{}) bool {
		t, ok := args[0].(*enemy.Enemy)
		if !ok {
			return false
		}
		c.a1Handle(t)
		return false
	}, "citlali-a1-frozen")
}

func (c *char) a1Handle(trg *enemy.Enemy) {
	if c.Base.Ascension < 1 {
		return
	}
	if !c.StatusIsActive(a1IcdKey) && c.nightsoulState.HasBlessing() && c.StatusIsActive(SkillKey) {
		c.nightsoulState.GeneratePoints(16)
		if c.Base.Cons >= 1 {
			c.Tags[SkillKey] += 3
		}
		c.AddStatus(a1IcdKey, 8*60, true)
	}
	shred := -0.2
	if c.Base.Cons >= 2 {
		shred = -0.4
	}
	trg.AddResistMod(combat.ResistMod{
		Base:  modifier.NewBaseWithHitlag("citlali-a1-shred-hydro", a1Duration),
		Ele:   attributes.Hydro,
		Value: shred,
	})
	trg.AddResistMod(combat.ResistMod{
		Base:  modifier.NewBaseWithHitlag("citlali-a1-shred-pyro", a1Duration),
		Ele:   attributes.Pyro,
		Value: shred,
	})
}

func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	c.Core.Events.Subscribe(event.OnNightsoulBurst, func(args ...interface{}) bool {
		if c.nightsoulState.HasBlessing() {
			c.nightsoulState.GeneratePoints(4)
		}
		return false
	}, "ciitlali-a4-nightsoul")
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)
		if ae.Info.ActorIndex != c.Index {
			return false
		}
		if ae.Info.Abil == "Dawnfrost Darkstar: Frostfall Storm DMG (DoT)" {
			ae.Info.FlatDmg += 0.9 * c.Stat(attributes.EM)
		}
		if ae.Info.Abil == "Edict of Entwined Splendor: Ice Storm DMG" {
			ae.Info.FlatDmg += 12 * c.Stat(attributes.EM)
		}
		return false
	}, "citlali-a1-melt")
}
