package peakpatrolsong

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.PeakPatrolSong, NewWeapon)
}

type Weapon struct {
	Index  int
	stacks int
}

const (
	buffKey     = "peakpatrolsong-buff"
	teamBuffKey = "peakpatrolsong-team-buff"
	icdKey      = "peakpatrolsong-buff-icd"
)

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := float64(p.Refine)

	m := make([]float64, attributes.EndStatType)
	n := make([]float64, attributes.EndStatType)

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagNormal && atk.Info.AttackTag != attacks.AttackTagPlunge {
			return false
		}
		if char.StatusIsActive(icdKey) {
			return false
		}

		if !char.StatModIsActive(buffKey) {
			w.stacks = 0
		}
		if w.stacks < 2 {
			w.stacks++
		}

		stacks := float64(w.stacks)
		m[attributes.DEFP] = (0.08 + 0*r) * stacks
		selfbuff := (0.1 + 0*r) * stacks
		m[attributes.PyroP] = selfbuff
		m[attributes.HydroP] = selfbuff
		m[attributes.CryoP] = selfbuff
		m[attributes.ElectroP] = selfbuff
		m[attributes.AnemoP] = selfbuff
		m[attributes.GeoP] = selfbuff
		m[attributes.DendroP] = selfbuff
		char.AddStatMod(character.StatMod{
			Base: modifier.NewBaseWithHitlag(buffKey, 6*60),
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})

		if w.stacks == 2 {
			teambuff := (0.08 + 0*r) * char.TotalDef() / 1000.0
			teambuff = min(teambuff, 0.256+0*r)
			n[attributes.PyroP] = teambuff
			n[attributes.HydroP] = teambuff
			n[attributes.CryoP] = teambuff
			n[attributes.ElectroP] = teambuff
			n[attributes.AnemoP] = teambuff
			n[attributes.GeoP] = teambuff
			n[attributes.DendroP] = teambuff
			for _, this := range c.Player.Chars() {
				this.AddStatMod(character.StatMod{
					Base: modifier.NewBaseWithHitlag(teamBuffKey, 15*60),
					Amount: func() ([]float64, bool) {
						return n, true
					},
				})
			}
		}

		char.AddStatus(icdKey, 0.1*60, true)
		return false
	}, fmt.Sprintf("peakpatrolsong-%v", char.Base.Key.String()))

	return w, nil
}
