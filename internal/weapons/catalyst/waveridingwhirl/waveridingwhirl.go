package waveridingwhirl

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

const (
	icdKey = "waveridingwhirl-icd"
)

func init() {
	core.RegisterWeaponFunc(keys.WaveRidingWhirl, NewWeapon)
}

type Weapon struct {
	Index   int
	core    *core.Core
	counter int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error {
	w.counter = 0
	for _, x := range w.core.Player.Chars() {
		if x.Base.Element == attributes.Hydro {
			w.counter++
		}
	}

	w.counter = min(w.counter, 2)

	return nil
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	c.Events.Subscribe(event.OnSkill, func(args ...interface{}) bool {
		if c.Player.Active() != char.Index {
			return false
		}
		if char.StatusIsActive(icdKey) {
			return false
		}

		char.AddStatus(icdKey, 15*60, true)

		mHP := make([]float64, attributes.EndStatType)
		mHP[attributes.HPP] = 0.2 + float64(r)*0 + (0.12+float64(r)*0)*float64(w.counter) //refine placeholder
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("waveridingwhirl-hp%", 10*60),
			AffectedStat: attributes.HPP,
			Amount: func() ([]float64, bool) {
				return mHP, true
			},
		})

		return false
	}, fmt.Sprintf("waveridingwhirl-skill-%v", char.Base.Key.String()))

	return w, nil
}
