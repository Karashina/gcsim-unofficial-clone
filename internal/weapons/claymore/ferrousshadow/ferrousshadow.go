package ferrousshadow

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
	core.RegisterWeaponFunc(keys.FerrousShadow, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// HPが70/75/80/85/90%以下になると、重撃ダメージが30/35/40/45/50%増加し、
// 重撃が中断されにくくなる。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = 0.25 + float64(r)*0.05
	hpCheck := 0.65 + float64(r)*0.05

	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("ferrousshadow", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			// 重撃でなければバフを適用しない
			if atk.Info.AttackTag != attacks.AttackTagExtra {
				return nil, false
			}
			// HP閾値を超えている場合はバフを適用しない
			if char.CurrentHPRatio() > hpCheck {
				return nil, false
			}
			return m, true
		},
	})

	return w, nil
}
