package starcallerswatch

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/core/player/shield"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.StarcallersWatch, NewWeapon)
}

const (
	IcdKey = "starcallerswatch-icd"
)

type Weapon struct {
	Index int
	core  *core.Core
	char  *character.CharWrapper
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		core: c,
		char: char,
	}
	r := p.Refine

	em := 100 + float64(r)*0

	//free em
	m := make([]float64, attributes.EndStatType)
	w.char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("starcallerswatch-self", -1),
		AffectedStat: attributes.NoStat,
		Amount: func() ([]float64, bool) {
			m[attributes.EM] = em
			return m, true
		},
	})

	c.Events.Subscribe(event.OnShielded, func(args ...interface{}) bool {
		shd := args[0].(shield.Shield)

		if shd.ShieldOwner() != char.Index {
			return false
		}

		if char.StatusIsActive(IcdKey) {
			return false
		}

		for _, chr := range c.Player.Chars() {
			n := make([]float64, attributes.EndStatType)
			chr.AddAttackMod(character.AttackMod{
				Base: modifier.NewBaseWithHitlag("starcallerswatch-dmg", 15*60),
				Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
					if atk.Info.ActorIndex != c.Player.Active() {
						return nil, false
					} else {
						return n, true
					}
				},
			})
		}

		char.AddStatus(IcdKey, 14*60, true)

		return false
	}, fmt.Sprintf("starcallerswatch-sheild-%v", char.Base.Key.String()))

	return w, nil
}
