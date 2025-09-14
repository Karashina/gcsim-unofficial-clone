package etherlightspindlelute

import (
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
	core.RegisterWeaponFunc(keys.EtherlightSpindleLute, NewWeapon)
}

type Weapon struct {
	Index int
	core  *core.Core
	char  *character.CharWrapper
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		core: c,
		char: char,
	}
	r := p.Refine

	em := 75 + float64(r)*25

	c.Events.Subscribe(event.OnActionExec, func(args ...interface{}) bool {
		idx := args[0].(int)
		if idx != char.Index {
			return false
		}
		act := args[1].(action.Action)
		if act != action.ActionSkill {
			return false
		}
		const buffKey = "etherlightspindlelute-em"
		m := make([]float64, attributes.EndStatType)
		w.char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase(buffKey, 20*60),
			AffectedStat: attributes.NoStat,
			Amount: func() ([]float64, bool) {
				m[attributes.EM] = em
				return m, true
			},
		})
		return false
	}, "etherlightspindlelute")

	return w, nil
}
