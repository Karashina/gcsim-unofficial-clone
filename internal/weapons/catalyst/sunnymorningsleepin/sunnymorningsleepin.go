package sunnymorningsleepin

import (
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
	core.RegisterWeaponFunc(keys.SunnyMorningSleepIn, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := float64(p.Refine)

	onSwirl := make([]float64, attributes.EndStatType)
	onSwirl[attributes.EM] = 120 + 0*r
	onSkill := make([]float64, attributes.EndStatType)
	onSkill[attributes.EM] = 96 + 0*r
	onBurst := make([]float64, attributes.EndStatType)
	onBurst[attributes.EM] = 32 + 0*r

	swirlcb := func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)

		if ae.Info.ActorIndex != char.Index {
			return false
		}

		char.AddStatMod(character.StatMod{
			Base: modifier.NewBase("sunnymorningsleepin-onswirl", 6*60),
			Amount: func() ([]float64, bool) {
				return onSwirl, true
			},
		})

		return false
	}

	c.Events.Subscribe(event.OnSwirlCryo, swirlcb, "sunnymorningsleepin-onswirl")
	c.Events.Subscribe(event.OnSwirlElectro, swirlcb, "sunnymorningsleepin-onswirl")
	c.Events.Subscribe(event.OnSwirlHydro, swirlcb, "sunnymorningsleepin-onswirl")
	c.Events.Subscribe(event.OnSwirlPyro, swirlcb, "sunnymorningsleepin-onswirl")

	c.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)

		if ae.Info.ActorIndex != char.Index {
			return false
		}

		switch ae.Info.AttackTag {
		case attacks.AttackTagElementalArt:
			char.AddStatMod(character.StatMod{
				Base: modifier.NewBase("sunnymorningsleepin-onskill", 9*60),
				Amount: func() ([]float64, bool) {
					return onSkill, true
				},
			})
		case attacks.AttackTagElementalArtHold:
			char.AddStatMod(character.StatMod{
				Base: modifier.NewBase("sunnymorningsleepin-onskill", 9*60),
				Amount: func() ([]float64, bool) {
					return onSkill, true
				},
			})
		case attacks.AttackTagElementalBurst:
			char.AddStatMod(character.StatMod{
				Base: modifier.NewBase("sunnymorningsleepin-onburst", 30*60),
				Amount: func() ([]float64, bool) {
					return onBurst, true
				},
			})
		default:
			return false
		}
		return false
	}, "sunnymorningsleepin-onhit")

	return w, nil
}
