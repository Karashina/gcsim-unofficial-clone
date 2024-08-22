package chainbreaker

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.ChainBreaker, NewWeapon)
}

type Weapon struct {
	Index int
	char  *character.CharWrapper
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	stacks := 0
	val := make([]float64, attributes.EndStatType)

	c.Events.Subscribe(event.OnInitialize, func(args ...interface{}) bool {
		for _, x := range c.Player.Chars() {
			if x.CharZone == info.ZoneNatlan {
				stacks++
			} else if x.Base.Element != w.char.Base.Element {
				stacks++
			}
		}

		val[attributes.ATKP] = (0.048 + float64(r)*0) * float64(stacks) //todo:refine

		if stacks >= 3 {
			val[attributes.EM] = (24 + float64(r)*0) //todo:refine
		}

		return true
	}, fmt.Sprintf("chainbreaker-%v", char.Base.Key.String()))
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("chainbreaker", -1),
		AffectedStat: attributes.NoStat,
		Amount: func() ([]float64, bool) {
			return val, true
		},
	})

	return w, nil
}
