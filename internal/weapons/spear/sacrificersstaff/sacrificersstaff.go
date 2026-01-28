package sacrificersstaff

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
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.SacrificersStaff, NewWeapon)
}

type Weapon struct {
	Index  int
	stacks int
	c      *core.Core
	self   *character.CharWrapper
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{c: c, self: char}
	r := p.Refine

	atkPer := 0.06 + float64(r)*0.02  // 8/10/12/14/16 per refine
	erPer := 0.045 + float64(r)*0.015 // 6/7.5/9/10.5/12 per refine

	// Permanent stat mods that reference w.stacks
	mATK := make([]float64, attributes.EndStatType)
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("sacrificersstaff-atk", -1),
		AffectedStat: attributes.ATKP,
		Amount: func() ([]float64, bool) {
			mATK[attributes.ATKP] = atkPer * float64(w.stacks)
			return mATK, w.stacks > 0
		},
	})

	mER := make([]float64, attributes.EndStatType)
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("sacrificersstaff-er", -1),
		AffectedStat: attributes.ER,
		Amount: func() ([]float64, bool) {
			mER[attributes.ER] = erPer * float64(w.stacks)
			return mER, w.stacks > 0
		},
	})

	// On Elemental Skill hit -> add a stack (max 3), stacks last 6s each
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		// args: target, *combat.AttackEvent
		if _, ok := args[0].(*enemy.Enemy); !ok {
			return false
		}
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagElementalArt {
			return false
		}

		if w.stacks < 3 {
			w.stacks++
		}
		// schedule a decrement after 6s (per-stack timer)
		c.Tasks.Add(func() {
			if w.stacks > 0 {
				w.stacks--
			}
		}, 6*60)
		return false
	}, fmt.Sprintf("sacrificersstaff-%v", char.Base.Key.String()))

	return w, nil
}
