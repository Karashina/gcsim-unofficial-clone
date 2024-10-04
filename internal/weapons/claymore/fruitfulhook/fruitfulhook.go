package fruitfulhook

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.FruitfulHook, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	c.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)
		if ae.Info.AttackTag != attacks.AttackTagPlunge {
			return false
		}
		if c.Player.Active() != char.Index {
			return false
		}
		m := make([]float64, attributes.EndStatType)
		m[attributes.DmgP] = 0.16 + 0*float64(r)
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("fruitfulhook-additional", 10*60),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				switch atk.Info.AttackTag {
				case attacks.AttackTagNormal, attacks.AttackTagExtra, attacks.AttackTagPlunge:
					return m, true
				}
				return nil, false
			},
		})
		return false
	}, "fruitfulhook-plunge")

	n := make([]float64, attributes.EndStatType)
	n[attributes.CR] = 0.16 + 0*float64(r)
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("fruitfulhook-base", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			switch atk.Info.AttackTag {
			case attacks.AttackTagPlunge:
				return n, true
			}
			return nil, false
		},
	})

	return w, nil
}
