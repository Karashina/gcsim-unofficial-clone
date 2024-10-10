package xilonen

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	a1ICDKey       = "xilonen-a1-icd"
	a1NightsoulKey = "xilonen-a1-nightsoul"
	a4IcdKey       = "xilonen-a4-icd"
)

func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}

	switch c.A1Mode {
	case 1:
		c.a1buff = make([]float64, attributes.EndStatType)
		c.a1buff[attributes.DmgP] = 0.3
		c.AddAttackMod(character.AttackMod{
			Base: modifier.NewBaseWithHitlag("xilonen-a1-na-buff", -1),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				return c.a1buff, c.StatusIsActive(skillKey) && atk.Info.AttackTag == attacks.AttackTagNormal || atk.Info.AttackTag == attacks.AttackTagPlunge
			},
		})
	case 2:
		c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
			ae := args[1].(*combat.AttackEvent)

			if !c.StatusIsActive(skillKey) {
				return false
			}
			if ae.Info.ActorIndex != c.Index {
				return false
			}
			if ae.Info.AttackTag != attacks.AttackTagNormal && ae.Info.AttackTag != attacks.AttackTagPlunge {
				return false
			}
			if c.StatusIsActive(a1ICDKey) {
				return false
			}

			c.AddNightsoul(a1NightsoulKey, 35)
			c.AddStatus(a1ICDKey, 6, true)

			return false
		}, "xilonen-a1")
	default:
		c.Core.Log.NewEvent("Invalid A1 Mode!", glog.LogCharacterEvent, c.Index)
		return
	}
}

func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	c.Core.Events.Subscribe(event.OnNightsoulBurst, func(args ...interface{}) bool {
		c.a4buff = make([]float64, attributes.EndStatType)
		c.a4buff[attributes.DEFP] = 0.20
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("xilonen-a4", 15*60),
			AffectedStat: attributes.DEFP,
			Amount: func() ([]float64, bool) {
				return c.a4buff, true
			},
		})
		return false
	}, "xilonen-a4")
}

func (c *char) a4AdditionalNsB() {
	if c.Base.Ascension < 4 {
		return
	}
	if c.StatusIsActive(a4IcdKey) {
		return
	}

	// fake ae for some reason
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "fake attack for xilonen a4",
	}
	ae := combat.AttackEvent{
		Info:        ai,
		Pattern:     combat.NewSingleTargetHit(1),
		SourceFrame: c.Core.F,
	}
	c.Core.Log.NewEvent("Xilonen A4 Nightsoul Burst", glog.LogCharacterEvent, c.Index)
	c.Core.Events.Emit(event.OnNightsoulBurst, nil, ae)
	c.AddStatus(a4IcdKey, 14*60, true)
}
