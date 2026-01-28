package crimsonmoonssemblance

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	icdKey      = "crimsonmoonssemblance-icd"
	icdDuration = 14 * 60
)

func init() {
	core.RegisterWeaponFunc(keys.CrimsonMoonsSemblance, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	var w Weapon
	refine := p.Refine

	c.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)

		if ae.Info.AttackTag != attacks.AttackTagExtra {
			return false
		}

		if c.Player.Active() != char.Index {
			return false
		}

		if char.StatusIsActive(icdKey) {
			return false
		}

		char.AddStatus(icdKey, icdDuration, true)
		char.ModifyHPDebtByRatio(0.25)

		return false
	}, fmt.Sprintf("crimsonmoonssemblance-hit-%v", char.Base.Key.String()))

	m := make([]float64, attributes.EndStatType)

	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag("crimsonmoonssemblance-bonus", -1),
		AffectedStat: attributes.DmgP,
		Amount: func() ([]float64, bool) {
			maxhp := char.MaxHP()
			m[attributes.DmgP] = 0.0
			if char.CurrentHPDebt() > 0 {
				m[attributes.DmgP] += 0.08 + 0.04*float64(refine)
			}
			if char.CurrentHPDebt() >= 0.3*maxhp {
				m[attributes.DmgP] += 0.16 + 0.08*float64(refine)
			}
			return m, true
		},
	})

	return &w, nil
}
