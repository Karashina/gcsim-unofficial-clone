package songofstillness

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.SongOfStillness, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// After the wielder is healed, they will deal 16/20/24/28/32% more DMG for 8s.
// This can be triggered even when the character is not on the field.
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	dmg := 0.12 + float64(r)*0.04
	duration := 8 * 60
	val := make([]float64, attributes.EndStatType)
	val[attributes.DmgP] = dmg
	c.Events.Subscribe(event.OnHeal, func(args ...interface{}) bool {
		index := args[1].(int)
		if index != char.Index {
			return false
		}
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("songofstillness-dmg-boost", duration),
			AffectedStat: attributes.DmgP,
			Amount: func() ([]float64, bool) {
				return val, true
			},
		})
		return false
	}, fmt.Sprintf("songofstillness-%v", char.Base.Key.String()))
	return w, nil
}

