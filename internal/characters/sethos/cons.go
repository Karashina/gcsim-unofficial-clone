package sethos

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/core/targets"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	c2aimkey    = "sethos-c2-aim"
	c2energykey = "sethos-c2-energy"
	c2burstkey  = "sethos-c2-energy"
	c6icdkey    = "sethos-c6-icd"
)

func (c *char) c1() {
	m := make([]float64, attributes.EndStatType)
	m[attributes.CR] = 0.15
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("sethos-c1", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.Abil != "Shadowpiercing Shot DMG" {
				return nil, false
			}
			return m, true
		},
	})
}

func (c *char) c2() {

	c.c2stacks = nTrue(c.StatusIsActive(c2aimkey), c.StatusIsActive(c2energykey), c.StatusIsActive(c2burstkey))

	if c.c2stacks >= 2 {
		c.c2stacks = 2
	}
	m := make([]float64, attributes.EndStatType)

	m[attributes.DmgP] = 0.3
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("sethos-c2", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.ActorIndex != c.Index {
				return nil, false
			}
			return m, true
		},
	})
}

func (c *char) c4cb() combat.AttackCBFunc {
	if c.Base.Cons < 4 {
		return nil
	}
	c.c4count = 0
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if a.AttackEvent.Info.AttackTag != attacks.AttackTagExtra {
			return
		}
		c.c4count++
		if c.c4count == 2 {
			dur := 10 * 60
			for _, char := range c.Core.Player.Chars() {
				char.AddStatMod(character.StatMod{
					Base:         modifier.NewBaseWithHitlag("sethos-c4", dur),
					AffectedStat: attributes.EM,
					Amount: func() ([]float64, bool) {
						return c.c4Buff, true
					},
				})
			}
			c.Core.Log.NewEvent("sethos c4 triggered", glog.LogCharacterEvent, c.Index).Write("em snapshot", c.c4Buff[attributes.EM]).Write("expiry", c.Core.F+dur)
			c.c4count = 0
		}
	}
}

func nTrue(b ...bool) int {
	n := 0
	for _, v := range b {
		if v {
			n++
		}
	}
	return n
}
