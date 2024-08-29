package earthshaker

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/gadget"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.EarthShaker, NewWeapon)
}

const (
	buffKey = "earthshaker-buff"
)

type Weapon struct {
	core   *core.Core
	char   *character.CharWrapper
	refine int
	buffE  []float64
	Index  int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		core:   c,
		char:   char,
		refine: p.Refine,
		buffE:  make([]float64, attributes.EndStatType),
	}

	addBuff := func(args ...interface{}) bool {
		if _, ok := args[0].(*gadget.Gadget); ok {
			return false
		}

		w.char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBaseWithHitlag(buffKey, 8*60),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				w.buffE[attributes.DmgP] = 0.12 + 0.04*float64(w.refine)
				switch atk.Info.AttackTag {
				case attacks.AttackTagElementalArt:
					return w.buffE, true
				case attacks.AttackTagElementalArtHold:
					return w.buffE, true
				default:
					return nil, false
				}
			},
		})

		return false
	}

	subKey := "earthshaker-" + char.Base.Key.String()

	c.Events.Subscribe(event.OnVaporize, addBuff, subKey)
	c.Events.Subscribe(event.OnMelt, addBuff, subKey)
	c.Events.Subscribe(event.OnOverload, addBuff, subKey)
	c.Events.Subscribe(event.OnSwirlPyro, addBuff, subKey)
	c.Events.Subscribe(event.OnCrystallizePyro, addBuff, subKey)
	c.Events.Subscribe(event.OnBurning, addBuff, subKey)
	c.Events.Subscribe(event.OnBurgeon, addBuff, subKey)

	return w, nil
}
