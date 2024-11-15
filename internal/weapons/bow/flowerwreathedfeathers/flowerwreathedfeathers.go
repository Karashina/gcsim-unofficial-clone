package flowerwreathedfeathers

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
	core.RegisterWeaponFunc(keys.FlowerWreathedFeathers, NewWeapon)
}

type Weapon struct {
	Index        int
	globalstacks int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	dmgpers := 0.06 + 0*float64(r) //refine placeholder

	m := make([]float64, attributes.EndStatType)

	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("flowerwreathedfeathers", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagNormal && atk.Info.AttackTag != attacks.AttackTagExtra {
				return nil, false
			}
			m[attributes.DmgP] = 0
			travel := float64(atk.Snapshot.SourceFrame-atk.SourceFrame) / 60
			stacks := int(travel / 0.5)
			w.globalstacks += stacks
			if w.globalstacks > 6 {
				w.globalstacks = 6
			}
			m[attributes.DmgP] += dmgpers * float64(w.globalstacks)
			return m, true
		},
	})

	c.Events.Subscribe(event.OnTick, func(args ...interface{}) bool {

		if !char.StatusIsActive("flowerwreathedfeathers-check") {
			w.globalstacks = 0
		}

		return false
	}, fmt.Sprintf("flowerwreathedfeathers-reset-%v", char.Base.Key.String()))

	c.Events.Subscribe(event.OnActionExec, func(args ...interface{}) bool {
		idx := args[0].(int)
		act := args[1].(action.Action)
		if idx != char.Index {
			return false
		}
		if c.Player.Active() != char.Index {
			return false
		}
		if act != action.ActionAim {
			return false
		}

		char.AddStatus("flowerwreathedfeathers-check", 10*60, true)

		return false
	}, fmt.Sprintf("flowerwreathedfeathers-ca-%v", char.Base.Key.String()))

	return w, nil
}
