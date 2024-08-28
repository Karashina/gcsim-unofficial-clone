package scrolloftheheroofcindercity

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/gadget"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterSetFunc(keys.ScrollOfTheHeroOfCinderCity, NewSet)
}

type Set struct {
	Index int
	Count int
}

func (s *Set) SetIndex(idx int) { s.Index = idx }
func (s *Set) GetCount() int    { return s.Count }
func (s *Set) Init() error      { return nil }

func NewSet(c *core.Core, char *character.CharWrapper, count int, param map[string]int) (info.Set, error) {
	s := Set{Count: count}
	m := make([]float64, attributes.EndStatType)
	dmg := 0.12

	if count >= 2 {
		c.Events.Subscribe(event.OnNightsoulBurst, func(args ...interface{}) bool {
			char.AddEnergy("scrolloftheheroofcindercity", 6)
			return false
		}, "scroll-2pc")
	}
	if count >= 4 {
		scrollfunc := func(ele1 attributes.Element, ele2 attributes.Element, key string, gadgetEmit bool) func(args ...interface{}) bool {
			icd := -1
			return func(args ...interface{}) bool {
				if _, ok := args[0].(*gadget.Gadget); ok != gadgetEmit {
					return false
				}
				ae := args[1].(*combat.AttackEvent)

				if ae.Info.ActorIndex != char.Index {
					return false
				}
				if c.F < icd {
					return false
				}
				icd = c.F + 1

				for _, char := range c.Player.Chars() {
					if char.OnNightsoul {
						char.AddStatMod(character.StatMod{
							Base:         modifier.NewBaseWithHitlag("scroll-4pc-buff", 20*60),
							AffectedStat: attributes.NoStat,
							Amount: func() ([]float64, bool) {
								m[attributes.PyroP] = 0
								m[attributes.HydroP] = 0
								m[attributes.CryoP] = 0
								m[attributes.ElectroP] = 0
								m[attributes.AnemoP] = 0
								m[attributes.GeoP] = 0
								m[attributes.DendroP] = 0
								m[attributes.EleToDmgP(ele1)] = dmg + 0.16
								if ele2 != attributes.NoElement {
									m[attributes.EleToDmgP(ele2)] = dmg + 0.16
								}
								return m, true
							},
						})
					} else {
						char.AddStatMod(character.StatMod{
							Base:         modifier.NewBaseWithHitlag("scroll-4pc-buff", 15*60),
							AffectedStat: attributes.NoStat,
							Amount: func() ([]float64, bool) {
								m[attributes.PyroP] = 0
								m[attributes.HydroP] = 0
								m[attributes.CryoP] = 0
								m[attributes.ElectroP] = 0
								m[attributes.AnemoP] = 0
								m[attributes.GeoP] = 0
								m[attributes.DendroP] = 0
								m[attributes.EleToDmgP(ele1)] = dmg
								if ele2 != attributes.NoElement {
									m[attributes.EleToDmgP(ele2)] = dmg
								}
								return m, true
							},
						})
					}
				}
				c.Log.NewEvent("scroll 4pc proc'd", glog.LogWeaponEvent, char.Index).
					Write("trigger", key)
				return false
			}
		}

		switch char.Base.Element {
		case attributes.Anemo:
			c.Events.Subscribe(event.OnSwirlCryo, scrollfunc(attributes.Anemo, attributes.Cryo, "scroll-swirlcryo", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnSwirlElectro, scrollfunc(attributes.Anemo, attributes.Electro, "scroll-swirlelectro", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnSwirlHydro, scrollfunc(attributes.Anemo, attributes.Hydro, "scroll-swirlelectro", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnSwirlPyro, scrollfunc(attributes.Anemo, attributes.Pyro, "scroll-swirlelectro", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
		case attributes.Cryo:
			c.Events.Subscribe(event.OnSwirlCryo, scrollfunc(attributes.Anemo, attributes.Cryo, "scroll-swirlcryo", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnSuperconduct, scrollfunc(attributes.Electro, attributes.Cryo, "scroll-superconduct", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnCrystallizeCryo, scrollfunc(attributes.Geo, attributes.Cryo, "scroll-crystallizecryo", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnFrozen, scrollfunc(attributes.Hydro, attributes.Cryo, "scroll-frozen", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnMelt, scrollfunc(attributes.Pyro, attributes.Cryo, "scroll-melt", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
		case attributes.Dendro:
			c.Events.Subscribe(event.OnQuicken, scrollfunc(attributes.Dendro, attributes.Electro, "scroll-quicken", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnSpread, scrollfunc(attributes.Dendro, attributes.NoElement, "scroll-aggravate", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnBloom, scrollfunc(attributes.Dendro, attributes.Hydro, "scroll-bloom", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnBurning, scrollfunc(attributes.Dendro, attributes.Pyro, "scroll-burning", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
		case attributes.Electro:
			c.Events.Subscribe(event.OnSwirlElectro, scrollfunc(attributes.Anemo, attributes.Electro, "scroll-swirlelectro", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnSuperconduct, scrollfunc(attributes.Electro, attributes.Cryo, "scroll-superconduct", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnQuicken, scrollfunc(attributes.Dendro, attributes.Electro, "scroll-quicken", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnAggravate, scrollfunc(attributes.Electro, attributes.NoElement, "scroll-aggravate", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnHyperbloom, scrollfunc(attributes.Electro, attributes.NoElement, "scroll-hyperbloom", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnCrystallizeElectro, scrollfunc(attributes.Geo, attributes.Electro, "scroll-crystallizeelectro", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnElectroCharged, scrollfunc(attributes.Hydro, attributes.Electro, "scroll-electrocharged", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnOverload, scrollfunc(attributes.Pyro, attributes.Electro, "scroll-overload", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
		case attributes.Geo:
			c.Events.Subscribe(event.OnCrystallizeCryo, scrollfunc(attributes.Geo, attributes.Cryo, "scroll-crystallizecryo", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnCrystallizeElectro, scrollfunc(attributes.Geo, attributes.Electro, "scroll-crystallizeelectro", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnCrystallizeHydro, scrollfunc(attributes.Geo, attributes.Hydro, "scroll-crystallizehydro", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnCrystallizePyro, scrollfunc(attributes.Geo, attributes.Pyro, "scroll-crystallizepyro", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
		case attributes.Hydro:
			c.Events.Subscribe(event.OnSwirlHydro, scrollfunc(attributes.Anemo, attributes.Hydro, "scroll-swirlelectro", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnFrozen, scrollfunc(attributes.Hydro, attributes.Cryo, "scroll-frozen", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnBloom, scrollfunc(attributes.Dendro, attributes.Hydro, "scroll-bloom", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnElectroCharged, scrollfunc(attributes.Hydro, attributes.Electro, "scroll-electrocharged", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnCrystallizeHydro, scrollfunc(attributes.Geo, attributes.Hydro, "scroll-crystallizehydro", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnVaporize, scrollfunc(attributes.Hydro, attributes.Pyro, "scroll-vaporize", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
		case attributes.Pyro:
			c.Events.Subscribe(event.OnSwirlPyro, scrollfunc(attributes.Anemo, attributes.Pyro, "scroll-swirlelectro", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnMelt, scrollfunc(attributes.Pyro, attributes.Cryo, "scroll-melt", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnBurning, scrollfunc(attributes.Dendro, attributes.Pyro, "scroll-burning", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnBurgeon, scrollfunc(attributes.Pyro, attributes.NoElement, "scroll-burning", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnOverload, scrollfunc(attributes.Pyro, attributes.Electro, "scroll-overload", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnCrystallizePyro, scrollfunc(attributes.Geo, attributes.Pyro, "scroll-crystallizepyro", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnVaporize, scrollfunc(attributes.Hydro, attributes.Pyro, "scroll-vaporize", true), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
		default:
		}
	}
	return &s, nil
}
