package mizuki

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const c1Key = "mizuki-c1"

func (c *char) c1check(src int) func() {
	return func() {
		if c.Base.Cons < 1 {
			return
		}
		if c.dreamdrifterSrc != src {
			return
		}
		if !c.StatusIsActive(skillKey) {
			return
		}
		enemies := c.Core.Combat.RandomEnemiesWithinArea(combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 7), nil, 99)
		for _, enemy := range enemies {
			enemy.AddStatus(c1Key, 3*60, true)
		}
		c.QueueCharTask(c.c1check(src), 3.5*60)
	}
}

func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		t, ok := args[0].(*enemy.Enemy)
		ae := args[1].(*combat.AttackEvent)
		if !ok {
			return false
		}
		if !t.StatusIsActive(c1Key) {
			return false
		}
		switch ae.Info.AttackTag {
		case attacks.AttackTagSwirlCryo:
		case attacks.AttackTagSwirlElectro:
		case attacks.AttackTagSwirlHydro:
		case attacks.AttackTagSwirlPyro:
		default:
			return false
		}
		ae.Info.FlatDmg += 11 * c.Stat(attributes.EM)
		t.DeleteStatus(c1Key)
		return false
	}, "mizuki-c1")
}

func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}
	m := make([]float64, attributes.EndStatType)

	var elems = []attributes.Stat{attributes.PyroP, attributes.HydroP, attributes.CryoP, attributes.ElectroP}
	// recalc em
	dmg := 0.0004 * c.NonExtraStat(attributes.EM)
	length := c.StatusDuration(skillKey)

	for _, ele := range elems {
		for _, char := range c.Core.Player.Chars() {
			char.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag(fmt.Sprintf("mizuki-c2-%v", ele), length),
				AffectedStat: ele,
				Extra:        true,
				Amount: func() ([]float64, bool) {
					clear(m)
					m[ele] = dmg
					return m, true
				},
			})
		}
	}
}

func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}
	if c.c4Count > 4 {
		return
	}
	c.c4Count++
	c.AddEnergy("mizuki-c4", 5)
}

func (c *char) c6() {
	if c.Base.Cons < 6 {
		return
	}
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		if !c.StatusIsActive(skillKey) {
			return false
		}
		ae := args[1].(*combat.AttackEvent)

		switch ae.Info.AttackTag {
		case attacks.AttackTagSwirlCryo:
		case attacks.AttackTagSwirlElectro:
		case attacks.AttackTagSwirlHydro:
		case attacks.AttackTagSwirlPyro:
		default:
			return false
		}

		//TODO: should this really be +=??
		ae.Snapshot.Stats[attributes.CR] += 0.3
		ae.Snapshot.Stats[attributes.CD] += 1

		c.Core.Log.NewEvent("mizuki c6 buff", glog.LogCharacterEvent, ae.Info.ActorIndex).
			Write("final_crit", ae.Snapshot.Stats[attributes.CR])

		return false
	}, "mizuki-c6")
}
