package ezpitzal

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.FluteOfEzpitzal, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	defp := 0.16 + float64(r)*0 //todo: refines
	val := make([]float64, attributes.EndStatType)

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

		val[attributes.DEFP] = defp
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("ezpitzal-defp", 15*60),
			AffectedStat: attributes.DEFP,
			Amount: func() ([]float64, bool) {
				return val, true
			},
		})
		return false
	}, fmt.Sprintf("ezpitzal-skill-%v", char.Base.Key.String()))

	return w, nil
}
