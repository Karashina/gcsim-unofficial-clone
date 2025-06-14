package azurelight

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
	core.RegisterWeaponFunc(keys.AzureLight, NewWeapon)
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
		if c.Player.Active() != char.Index {
			return false
		}
		if act != action.ActionSkill {
			return false
		}

		m := make([]float64, attributes.EndStatType)
		m[attributes.ATKP] = 0.24 + 0*float64(r)

		mNoEnergy := make([]float64, attributes.EndStatType)
		mNoEnergy[attributes.ATKP] = 0.48 + 0*float64(r)
		mNoEnergy[attributes.CD] = 0.40 + 0*float64(r)

		// refresh mod
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("azulelight", 12*60),
			AffectedStat: attributes.NoStat,
			Amount: func() ([]float64, bool) {
				if char.Energy <= 0 {
					return mNoEnergy, true
				} else {
					return m, true
				}
			},
		})

		return false
	}, fmt.Sprintf("azulelight-%v", char.Base.Key.String()))
	return w, nil
}
