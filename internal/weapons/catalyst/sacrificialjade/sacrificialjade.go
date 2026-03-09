package sacrificialjade

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.SacrificialJade, NewWeapon)
}

type Weapon struct {
	Index    int
	refine   int
	c        *core.Core
	char     *character.CharWrapper
	lastSwap int
	buff     []float64
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }

func (w *Weapon) Init() error {
	if w.c.Player.Active() != w.char.Index {
		w.lastSwap = w.c.F
		w.c.Tasks.Add(w.getBuffs(w.lastSwap), 5*60)
	}
	return nil
}

// フィールドに5秒以上いない時、HP上限が32%増加し、元素熟知が40増加する。
// これらの効果は装備者がフィールドに10秒以上いた後に解除される。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		refine:   p.Refine,
		c:        c,
		char:     char,
		lastSwap: -1,
		buff:     make([]float64, attributes.EndStatType),
	}
	if p.Params["stacks"] == 1 {
		w.addBuff()
	}

	w.char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("sacrificial-jade", -1),
		AffectedStat: attributes.NoStat,
		Amount: func() ([]float64, bool) {
			return w.buff, true
		},
	})

	c.Events.Subscribe(event.OnCharacterSwap, func(args ...interface{}) bool {
		prev := args[0].(int)
		next := args[1].(int)
		if prev == char.Index {
			w.lastSwap = c.F
			w.c.Tasks.Add(w.getBuffs(w.lastSwap), 5*60)
			return false
		}
		if next == char.Index {
			w.lastSwap = c.F
			w.c.Tasks.Add(w.clearBuffs(w.lastSwap), 10*60)
			return false
		}
		return false
	}, fmt.Sprintf("sacrificial-jade-%v", char.Base.Key.String()))

	return w, nil
}

func (w *Weapon) addBuff() {
	w.buff[attributes.HPP] = 0.24 + 0.08*float64(w.refine)
	w.buff[attributes.EM] = 30 + 10*float64(w.refine)
	w.c.Log.NewEvent("sacrificial jade gained buffs", glog.LogWeaponEvent, w.char.Index)
}

func (w *Weapon) getBuffs(src int) func() {
	return func() {
		if w.lastSwap != src {
			return
		}
		if w.c.Player.Active() == w.char.Index {
			return
		}
		w.addBuff()
	}
}

func (w *Weapon) clearBuffs(src int) func() {
	return func() {
		if w.lastSwap != src {
			return
		}
		if w.c.Player.Active() != w.char.Index {
			return
		}

		w.buff[attributes.HPP] = 0
		w.buff[attributes.EM] = 0
		w.c.Log.NewEvent("sacrificial jade lost buffs", glog.LogWeaponEvent, w.char.Index)
	}
}
