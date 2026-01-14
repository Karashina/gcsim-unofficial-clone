package ifa

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func (c *char) checkCurrentNightsoulPoint() {
	c.Core.Events.Subscribe(event.OnNightsoulGenerate, func(args ...interface{}) bool {
		index := args[0].(int)
		amt := args[1].(float64)
		c.charnightsoulpoints[index] += amt
		c.a1() //recalc a1
		return false
	}, "ifa-a1-nscheck-generate")
	c.Core.Events.Subscribe(event.OnNightsoulConsume, func(args ...interface{}) bool {
		index := args[0].(int)
		amt := args[1].(float64)
		c.charnightsoulpoints[index] -= amt
		c.a1() //recalc a1
		return false
	}, "ifa-a1-nscheck-consume")

}

func (c *char) totalPts() float64 {
	total := 0.0
	c2mult := 0.0
	if c.Base.Cons >= 2 {
		c2mult = 4.0
	}
	for i := 0; i < len(c.Core.Player.Chars()); i++ {
		if c.charnightsoulpoints[i] < 0.001 {
			continue
		}
		total += c.charnightsoulpoints[i]
	}
	total += max(0, (total-60)) * c2mult
	return total
}

func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	for _, chr := range c.Core.Player.Chars() {
		chr.AddLCReactBonusMod(character.LCReactBonusMod{
			Base: modifier.NewBase("ifa-a1-lc", 8*60),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				maxval := 150.0
				if c.Base.Cons >= 2 {
					maxval = 200.0
				}
				return min(maxval, c.totalPts()) * 0.002, false
			},
		})
		chr.AddReactBonusMod(character.ReactBonusMod{
			Base: modifier.NewBase("Field Medic's Vision (A1)", -1),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				if !c.nightsoulState.HasBlessing() {
					return 0, false
				}
				switch ai.AttackTag {
				case attacks.AttackTagSwirlCryo:
				case attacks.AttackTagSwirlElectro:
				case attacks.AttackTagSwirlPyro:
				case attacks.AttackTagSwirlHydro:
				case attacks.AttackTagECDamage:
				default:
					return 0, false
				}
				maxval := 150.0
				if c.Base.Cons >= 2 {
					maxval = 200.0
				}
				c.Core.Log.NewEvent(
					"Ifa a1 buff",
					glog.LogCharacterEvent,
					c.Index,
				).Write("Total NSP: ", min(maxval, c.totalPts())).Write("Buff: ", min(maxval, c.totalPts())*0.015)
				return min(maxval, c.totalPts()) * 0.015, false
			},
		})
	}
}

func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	c.Core.Events.Subscribe(event.OnNightsoulBurst, func(args ...interface{}) bool {
		m := make([]float64, attributes.EndStatType)
		m[attributes.EM] = 80
		c.AddStatMod(character.StatMod{
			Base: modifier.NewBaseWithHitlag("ifa-a4", 10*60),
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
		return false
	}, "ifa-a4")
}

