package halo

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
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
	core.RegisterWeaponFunc(keys.FracturedHalo, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	c.Events.Subscribe(event.OnActionExec, func(args ...interface{}) bool {
		idx := args[0].(int)
		act := args[1].(action.Action)

		if idx != char.Index {
			return false
		}

		if act != action.ActionSkill && act != action.ActionBurst {
			return false
		}

		mATK := make([]float64, attributes.EndStatType)
		mATK[attributes.ATKP] = 0.24 + float64(r)*0
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("halo-atk", 20*60),
			AffectedStat: attributes.ATKP,
			Amount: func() ([]float64, bool) {
				return mATK, true
			},
		})
		return false
	}, fmt.Sprintf("halo-atk-%v", char.Base.Key.String()))

	c.Events.Subscribe(event.OnShielded, func(args ...interface{}) bool {
		shd := args[0].(shield.Shield)

		if shd.ShieldOwner() != char.Index {
			return false
		}

		for _, v := range c.Player.Chars() {
			v.AddReactBonusMod(character.ReactBonusMod{
				Base: modifier.NewBase("halo-LCDMG", 20*60),
				Amount: func(ai combat.AttackInfo) (float64, bool) {
					switch ai.AttackTag {
					case attacks.AttackTagLCDamage:
					default:
						return 0, false
					}
					return 0.40 + float64(r)*0, false
				},
			})
		}
		return false
	}, fmt.Sprintf("halo-lcdmg-%v", char.Base.Key.String()))

	return w, nil
}
