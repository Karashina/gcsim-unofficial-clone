package blackmarrowlantern

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.BlackmarrowLantern, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	char.AddReactBonusMod(character.ReactBonusMod{
		Base: modifier.NewBaseWithHitlag("blackmarrowlantern-ec", -1),
		Amount: func(ai combat.AttackInfo) (float64, bool) {
			if ai.AttackTag != attacks.AttackTagBloom {
				return 0, false
			}
			return 0.36 + 0.12*float64(r), false
		},
	})
	char.AddLBReactBonusMod(character.LBReactBonusMod{
		Base: modifier.NewBaseWithHitlag("blackmarrowlantern-lb", -1),
		Amount: func(ai combat.AttackInfo) (float64, bool) {
			ascendant := false
			for _, chr := range c.Player.Chars() {
				if chr.MoonsignAscendant {
					ascendant = true
				}
			}
			if ascendant {
				return (0.09 + 0.03*float64(r)) * 2, false
			} else {
				return 0.09 + 0.03*float64(r), false
			}
		},
	})

	return w, nil
}
