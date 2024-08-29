package ringofyaxche

import (
	"fmt"

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
	core.RegisterWeaponFunc(keys.RingOfYaxche, NewWeapon)
}

const (
	buffKey = "yaxche-buff"
)

type Weapon struct {
	core   *core.Core
	char   *character.CharWrapper
	refine int
	buffNA []float64
	Index  int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		core:   c,
		char:   char,
		refine: p.Refine,
		buffNA: make([]float64, attributes.EndStatType),
	}

	c.Events.Subscribe(event.OnActionExec, func(args ...interface{}) bool {
		index := args[0].(int)
		e := args[1].(action.Action)
		if c.Player.Active() != index {
			return false
		}
		if char.Index != index {
			return false
		}
		if e != action.ActionSkill {
			return false
		}
		maxval := 0.12 + 0.04*float64(w.refine)
		buff := w.char.MaxHP() / 1000 * (0.005 + 0.001*float64(w.refine))

		w.char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBaseWithHitlag(buffKey, 10*60),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				w.buffNA[attributes.DmgP] = min(maxval, buff)
				switch atk.Info.AttackTag {
				case attacks.AttackTagNormal:
					return w.buffNA, true
				default:
					return nil, false
				}
			},
		})
		return false
	}, fmt.Sprintf("yaxche-skill-%v", char.Base.Key.String()))

	return w, nil
}
