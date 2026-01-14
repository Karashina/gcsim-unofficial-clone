package dawningfrost

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

func init() {
	core.RegisterWeaponFunc(keys.DawningFrost, NewWeapon)
}

type Weapon struct {
	Index int
	c     *core.Core
	self  *character.CharWrapper
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{c: c, self: char}
	r := p.Refine

	mCharge := 72 + float64(r-1)*18 // r1=72, r2=90, ... increments 18
	mSkill := 48 + float64(r-1)*12  // r1=48, r2=60, ... increments 12

	// Charged Attack hit -> EM for 10s
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		if _, ok := args[0].(*combat.Target); ok {
			// target might be of different type; just continue
		}
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if atk.Info.AttackTag == attacks.AttackTagExtra {
			mEM := make([]float64, attributes.EndStatType)
			mEM[attributes.EM] = mCharge
			char.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag("dawningfrost-charge", 10*60),
				AffectedStat: attributes.EM,
				Amount: func() ([]float64, bool) {
					return mEM, true
				},
			})
		}
		// Elemental Skill hit -> EM for 10s
		if atk.Info.AttackTag == attacks.AttackTagElementalArt {
			mEM := make([]float64, attributes.EndStatType)
			mEM[attributes.EM] = mSkill
			char.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag("dawningfrost-skill", 10*60),
				AffectedStat: attributes.EM,
				Amount: func() ([]float64, bool) {
					return mEM, true
				},
			})
		}
		return false
	}, fmt.Sprintf("dawningfrost-%v", char.Base.Key.String()))

	return w, nil
}

