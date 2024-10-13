package mountainbracingbolt

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
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
	core.RegisterWeaponFunc(keys.MountainBracingBolt, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	c.Events.Subscribe(event.OnActionExec, func(args ...interface{}) bool {
		act := args[1].(action.Action)
		if act != action.ActionSkill {
			return false
		}
		if c.Player.Active() == char.Index {
			return false
		}
		m := make([]float64, attributes.EndStatType)
		m[attributes.DmgP] = 0.09 + 0.03*float64(r)
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("mountainbracingbolt-additional", 8*60),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				switch atk.Info.AttackTag {
				case attacks.AttackTagElementalArt, attacks.AttackTagElementalArtHold:
					return m, true
				}
				return nil, false
			},
		})
		return false
	}, "mountainbracingbolt-skill")

	n := make([]float64, attributes.EndStatType)
	n[attributes.DmgP] = 0.09 + 0.03*float64(r)
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("mountainbracingbolt-base", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			switch atk.Info.AttackTag {
			case attacks.AttackTagElementalArt, attacks.AttackTagElementalArtHold:
				return n, true
			}
			return nil, false
		},
	})

	return w, nil
}
