package xilonen

import (
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	c4buffkey = "xilonen-c4-buff"
	c6buffKey = "xilonen-c6-buff"
	c6IcdKey  = "xilonen-c6-icd"
)

func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}
	i := 0
	geocount := 0
	for _, e := range c.SoundScapeSlot {
		if e == attributes.Geo {
			c.isSlotActive[i] = true
			geocount++
			c.Core.Log.NewEvent("Xilonen Geo Sample Activation from C2", glog.LogCharacterEvent, c.Index).
				Write("activated slot", i).
				Write("activated element", e)
		}
		i++
	}
	if geocount > 0 {
		c.Core.Tasks.Add(func() {
			enemies := c.Core.Combat.EnemiesWithinArea(combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 7.5), nil)
			for _, e := range enemies {
				e.AddResistMod(combat.ResistMod{
					Base:  modifier.NewBaseWithHitlag("xilonen-Geo", -1),
					Ele:   attributes.Geo,
					Value: skillRes[c.TalentLvlSkill()],
				})
			}
		}, 0)
	}
	chars := c.Core.Player.Chars()
	for _, this := range chars {
		switch this.Base.Element {
		case attributes.Pyro:
			for j := 0; j < 3; j++ {
				if c.SoundScapeSlot[j] == attributes.Pyro {
					c2buffPyro := make([]float64, attributes.EndStatType)
					c2buffPyro[attributes.ATKP] = 0.45
					this.AddStatMod(character.StatMod{
						Base:         modifier.NewBase("xilonen-c2-pyro", -1),
						AffectedStat: attributes.ATKP,
						Amount: func() ([]float64, bool) {
							return c2buffPyro, c.StatusIsActive(SampleActiveDurKey)
						},
					})
				}
			}
		case attributes.Hydro:
			for k := 0; k < 3; k++ {
				if c.SoundScapeSlot[k] == attributes.Hydro {
					c2buffHydro := make([]float64, attributes.EndStatType)
					c2buffHydro[attributes.HPP] = 0.45
					this.AddStatMod(character.StatMod{
						Base:         modifier.NewBase("xilonen-c2-hydro", -1),
						AffectedStat: attributes.HPP,
						Amount: func() ([]float64, bool) {
							return c2buffHydro, c.StatusIsActive(SampleActiveDurKey)
						},
					})
				}
			}
		case attributes.Cryo:
			for l := 0; l < 3; l++ {
				if c.SoundScapeSlot[l] == attributes.Cryo {
					c2buffCryo := make([]float64, attributes.EndStatType)
					c2buffCryo[attributes.CD] = 0.60
					this.AddStatMod(character.StatMod{
						Base:         modifier.NewBase("xilonen-c2-cryo", -1),
						AffectedStat: attributes.CD,
						Amount: func() ([]float64, bool) {
							return c2buffCryo, c.StatusIsActive(SampleActiveDurKey)
						},
					})
				}
			}
		case attributes.Electro:
			for m := 0; m < 3; m++ {
				if c.SoundScapeSlot[m] == attributes.Electro {
					this.AddEnergy("xilonen-c2-electro", 25)
					this.ReduceActionCooldown(action.ActionBurst, 6*60)
				}
			}
		case attributes.Geo:
			c2buffGeo := make([]float64, attributes.EndStatType)
			c2buffGeo[attributes.DmgP] = 0.50
			this.AddAttackMod(character.AttackMod{
				Base: modifier.NewBaseWithHitlag("xilonen-c2-geo", -1),
				Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
					return c2buffGeo, true
				},
			})
		default:
			continue
		}
	}
}

func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}

	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)

		if ae.Info.AttackTag != attacks.AttackTagNormal && ae.Info.AttackTag != attacks.AttackTagExtra && ae.Info.AttackTag != attacks.AttackTagPlunge {
			return false
		}
		char := c.Core.Player.ByIndex(ae.Info.ActorIndex)
		// do nothing if buff gone or burst count gone
		if char.Tags[c4buffkey] == 0 {
			return false
		}
		if !char.StatusIsActive(c4buffkey) {
			return false
		}

		dmgAdded := c.TotalDef() * 0.65
		ae.Info.FlatDmg += dmgAdded

		char.Tags[c4buffkey] -= 1

		c.Core.Log.NewEvent("Xilonen C4 adding damage", glog.LogPreDamageMod, ae.Info.ActorIndex).
			Write("damage_added", dmgAdded).
			Write("stacks_remaining_for_char", char.Tags[c4buffkey])

		return false
	}, "xilonen-c4")
}

func (c *char) c6setup() {
	if c.Base.Cons < 6 {
		return
	}
	c.AddStatus(c6buffKey, 5*60, true)
	c.AddStatus(c6IcdKey, 15*60, true)
	c.ExtendStatus(skillKey, 5*60)
	for i := 0; i <= 3; i++ {
		c.Core.Tasks.Add(func() {
			c.Core.Player.Heal(info.HealInfo{
				Caller:  c.Index,
				Target:  -1,
				Message: "Imperishable Night Carnival (C6)",
				Src:     c.TotalDef() * 1.2,
				Bonus:   c.Stat(attributes.Heal),
			})
		}, 89+89*i)
	}
	c.Core.Log.NewEvent("Xilonen C6 triggered", glog.LogPreDamageMod, c.Index)
}

func (c *char) c6() {
	if c.Base.Cons < 6 {
		return
	}
	c.Core.Events.Subscribe(event.OnActionExec, func(args ...interface{}) bool {
		a := args[1].(action.Action)

		if c.Core.Player.ActiveChar().Index != c.Index {
			return false
		}

		if a != action.ActionAttack && a != action.ActionHighPlunge && a != action.ActionDash && a != action.ActionJump {
			return false
		}

		if c.StatusIsActive(c6IcdKey) {
			return false
		}

		if !c.StatusIsActive(skillKey) {
			return false
		}
		c.c6setup()

		return false
	}, "xilonen-c6-check")
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)

		if ae.Info.ActorIndex != c.Index {
			return false
		}

		if !c.StatusIsActive(c6buffKey) {
			return false
		}

		if ae.Info.AttackTag != attacks.AttackTagNormal && ae.Info.AttackTag != attacks.AttackTagExtra && ae.Info.AttackTag != attacks.AttackTagPlunge {
			return false
		}

		dmgAdded := c.TotalDef() * 3
		ae.Info.FlatDmg += dmgAdded

		c.Core.Log.NewEvent("Xilonen C6 adding damage", glog.LogPreDamageMod, ae.Info.ActorIndex).
			Write("damage_added", dmgAdded)

		return false
	}, "xilonen-c6-addition")
}
