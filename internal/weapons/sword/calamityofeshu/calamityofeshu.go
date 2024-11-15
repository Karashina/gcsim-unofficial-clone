package calamityofeshu

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.CalamityOfEshu, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// While character is protected by a shield,
// DMG dealt by Normal and Charged Attacks is increased by 20%,
// and Normal and Charged Attacks CRIT Rate is increased by 8%.

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	m := make([]float64, attributes.EndStatType)
	m[attributes.CR] = 0.08 + 0*float64(r)     //refine placeholder
	m[attributes.DmgP] = 0.20 + 0.0*float64(r) //refine placeholder
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("calamityofeshu", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if c.Player.Shields.CharacterIsShielded(char.Index, c.Player.Active()) {
				switch atk.Info.AttackTag {
				case attacks.AttackTagNormal, attacks.AttackTagExtra:
					return m, true
				}
			}
			return nil, false
		},
	})

	return w, nil
}
