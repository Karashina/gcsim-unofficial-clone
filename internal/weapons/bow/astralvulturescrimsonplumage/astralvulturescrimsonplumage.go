package astralvulturescrimsonplumage

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
	core.RegisterWeaponFunc(keys.AstralVulturesCrimsonPlumage, NewWeapon)
}

type Weapon struct {
	Index int
	r     float64
	core  *core.Core
	char  *character.CharWrapper
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error {
	counter := 0
	for _, x := range w.core.Player.Chars() {
		if x.Base.Element != w.char.Base.Element {
			counter++
		}
	}

	counter = min(counter, 2)

	ca := make([]float64, attributes.EndStatType)
	burst := make([]float64, attributes.EndStatType)

	switch counter {
	case 0:
		ca[attributes.DmgP] = 0.0
		burst[attributes.DmgP] = 0.0
	case 1:
		ca[attributes.DmgP] = 0.15 + 0.05*w.r      //refine placeholder
		burst[attributes.DmgP] = 0.075 + 0.025*w.r //refine placeholder
	case 2:
		ca[attributes.DmgP] = 0.36 + 0.12*w.r    //refine placeholder
		burst[attributes.DmgP] = 0.18 + 0.06*w.r //refine placeholder
	default:
		ca[attributes.DmgP] = 0.0
		burst[attributes.DmgP] = 0.0
	}

	w.char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("astralvulturescrimsonplumage-ca", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagExtra {
				return nil, false
			}
			return ca, true
		},
	})

	w.char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("astralvulturescrimsonplumage-burst", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagElementalBurst {
				return nil, false
			}
			return burst, true
		},
	})

	return nil
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		core: c,
		char: char,
	}

	atkp := make([]float64, attributes.EndStatType)
	atkp[attributes.ATKP] = 0.18 + 0.06*w.r //refine placeholder

	for i := event.OnSwirlHydro; i <= event.OnSwirlPyro; i++ {
		c.Events.Subscribe(i, func(args ...interface{}) bool {
			atk := args[1].(*combat.AttackEvent)
			if atk.Info.ActorIndex != char.Index {
				return false
			}
			char.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag("astralvulturescrimsonplumage-atkp", 12*60),
				AffectedStat: attributes.ATKP,
				Amount: func() ([]float64, bool) {
					return atkp, true
				},
			})
			return false
		}, "astralvulturescrimsonplumage-swirl-"+char.Base.Key.String())
	}

	return w, nil
}
