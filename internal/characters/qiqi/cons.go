package qiqi

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func (c *char) c1(a combat.AttackCB) {
	e, ok := a.Target.(*enemy.Enemy)
	if !ok {
		return
	}

	if !e.StatusIsActive(talismanKey) {
		return
	}

	c.AddEnergy("qiqi-c1", 2)
	c.Core.Log.NewEvent("Qiqi C1 Activation - Adding 2 energy", glog.LogCharacterEvent, c.Index).
		Write("target", a.Target.Key())
}

// Qiqi's Normal and Charge Attack DMG against opponents affected by Cryo is increased by 15%.
func (c *char) c2() {
	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = .15
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("qiqi-c2", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagNormal && atk.Info.AttackTag != attacks.AttackTagExtra {
				return nil, false
			}

			e, ok := t.(*enemy.Enemy)
			if !ok {
				return nil, false
			}
			if !e.AuraContains(attributes.Cryo, attributes.Frozen) {
				return nil, false
			}

			return m, true
		},
	})
}

