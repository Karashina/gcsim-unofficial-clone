package halo

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/shield"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
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
		mATK[attributes.ATKP] = 0.18 + float64(r)*0.06
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
			v.AddLCReactBonusMod(character.LCReactBonusMod{
				Base: modifier.NewBase("halo-LCDMG", 20*60),
				Amount: func(ai combat.AttackInfo) (float64, bool) {
					return 0.30 + float64(r)*0.10, false
				},
			})
		}
		return false
	}, fmt.Sprintf("halo-lcdmg-%v", char.Base.Key.String()))

	return w, nil
}

