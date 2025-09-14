package prospectorsshovel

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.ProspectorsShovel, NewWeapon)
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
		Base: modifier.NewBaseWithHitlag("prospectorshovel-ec", -1),
		Amount: func(ai combat.AttackInfo) (float64, bool) {
			if ai.AttackTag != attacks.AttackTagECDamage {
				return 0, false
			}
			return 0.09 + 0.03*float64(r), false
		},
	})
	char.AddLCReactBonusMod(character.LCReactBonusMod{
		Base: modifier.NewBaseWithHitlag("prospectorshovel-lc", -1),
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
