package moonweaversdawn

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.MoonweaversDawn, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}

	r := p.Refine

	// Base burst damage bonus by refine level
	baseBonus := []float64{0.20, 0.25, 0.30, 0.35, 0.40}
	// Additional bonus when energy <= 40
	bonus40 := []float64{0.16, 0.20, 0.24, 0.28, 0.32}
	// Additional bonus when energy <= 60 (but > 40)
	bonus60 := []float64{0.28, 0.35, 0.42, 0.49, 0.56}

	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = baseBonus[r-1]

	if char.EnergyMax <= 40 {
		m[attributes.DmgP] += bonus40[r-1]
	} else if char.EnergyMax <= 60 {
		m[attributes.DmgP] += bonus60[r-1]
	}

	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("moonweaversdawn", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagElementalBurst {
				return nil, false
			}
			return m, true
		},
	})

	return w, nil
}
