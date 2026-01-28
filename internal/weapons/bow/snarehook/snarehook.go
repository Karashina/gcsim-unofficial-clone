package snarehook

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.SnareHook, NewWeapon)
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

	f := func(args ...interface{}) bool {
		mEM := make([]float64, attributes.EndStatType)

		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("snarehook-em", 12*60),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				ascendant := false
				for _, chr := range c.Player.Chars() {
					if chr.MoonsignAscendant {
						ascendant = true
					}
				}
				if ascendant {
					mEM[attributes.EM] = (45 + float64(r)*15) * 2
				} else {
					mEM[attributes.EM] = 45 + float64(r)*15
				}
				return mEM, true
			},
		})
		return false
	}

	for i := event.ReactionEventStartDelim + 1; i < event.OnShatter; i++ {
		w.core.Events.Subscribe(i, f, fmt.Sprintf("masterkey-%v", w.char.Base.Key.String()))
	}

	return w, nil
}
