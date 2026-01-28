package serenityscall

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
	core.RegisterWeaponFunc(keys.SerenityScall, NewWeapon)
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
		mHP := make([]float64, attributes.EndStatType)
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("serenityscall-hp", 12*60),
			AffectedStat: attributes.HPP,
			Amount: func() ([]float64, bool) {
				ascendant := false
				for _, chr := range c.Player.Chars() {
					if chr.MoonsignAscendant {
						ascendant = true
					}
				}
				if ascendant {
					mHP[attributes.HPP] = (0.12 + float64(r)*0.04) * 2
				} else {
					mHP[attributes.HPP] = 0.12 + float64(r)*0.04
				}
				return mHP, true
			},
		})
		return false
	}

	for i := event.ReactionEventStartDelim + 1; i < event.OnShatter; i++ {
		w.core.Events.Subscribe(i, f, fmt.Sprintf("serenityscall-%v", w.char.Base.Key.String()))
	}

	return w, nil
}
