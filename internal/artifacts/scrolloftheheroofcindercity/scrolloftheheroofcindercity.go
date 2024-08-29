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
	nanemo := make([]float64, attributes.EndStatType)
	ncryo := make([]float64, attributes.EndStatType)
	ndendro := make([]float64, attributes.EndStatType)
	nelectro := make([]float64, attributes.EndStatType)
	ngeo := make([]float64, attributes.EndStatType)
	nhydro := make([]float64, attributes.EndStatType)
	npyro := make([]float64, attributes.EndStatType)
	anemo := make([]float64, attributes.EndStatType)
	cryo := make([]float64, attributes.EndStatType)
	dendro := make([]float64, attributes.EndStatType)
	electro := make([]float64, attributes.EndStatType)
	geo := make([]float64, attributes.EndStatType)
	hydro := make([]float64, attributes.EndStatType)
	pyro := make([]float64, attributes.EndStatType)

	dmg := 0.12
	dmgnightsoul := 0.28

	if count >= 2 {
		c.Events.Subscribe(event.OnNightsoulBurst, func(args ...interface{}) bool {
			char.AddEnergy("scrolloftheheroofcindercity", 6)
			return false
		}, "scroll-2pc")
	}
	if count >= 4 {
		scrollfunc := func(ele1 attributes.Element, ele2 attributes.Element, key string) func(args ...interface{}) bool {
			return func(args ...interface{}) bool {

				ae := args[1].(*combat.AttackEvent)

				if ae.Info.ActorIndex != char.Index {
					return false
				}

				for _, x := range c.Player.Chars() {
					if char.OnNightsoul {
						switch ele1 {
						case attributes.Electro:
							x.AddStatMod(character.StatMod{
								Base:         modifier.NewBaseWithHitlag("scroll-4pc-nightsoul-electro", 20*60),
								AffectedStat: attributes.ElectroP,
								Amount: func() ([]float64, bool) {
									nelectro[attributes.ElectroP] = dmgnightsoul
									return nelectro, true
								},
							})
						case attributes.Pyro:
							x.AddStatMod(character.StatMod{
								Base:         modifier.NewBaseWithHitlag("scroll-4pc-nightsoul-pyro", 20*60),
								AffectedStat: attributes.PyroP,
								Amount: func() ([]float64, bool) {
									npyro[attributes.PyroP] = dmgnightsoul
									return npyro, true
								},
							})
						case attributes.Cryo:
							x.AddStatMod(character.StatMod{
								Base:         modifier.NewBaseWithHitlag("scroll-4pc-nightsoul-cryo", 20*60),
								AffectedStat: attributes.CryoP,
								Amount: func() ([]float64, bool) {
									ncryo[attributes.CryoP] = dmgnightsoul
									return ncryo, true
								},
							})
						case attributes.Hydro:
							x.AddStatMod(character.StatMod{
								Base:         modifier.NewBaseWithHitlag("scroll-4pc-nightsoul-hydro", 20*60),
								AffectedStat: attributes.HydroP,
								Amount: func() ([]float64, bool) {
									nhydro[attributes.HydroP] = dmgnightsoul
									return nhydro, true
								},
							})
						case attributes.Dendro:
							x.AddStatMod(character.StatMod{
								Base:         modifier.NewBaseWithHitlag("scroll-4pc-nightsoul-dendro", 20*60),
								AffectedStat: attributes.DendroP,
								Amount: func() ([]float64, bool) {
									ndendro[attributes.DendroP] = dmgnightsoul
									return ndendro, true
								},
							})
						case attributes.Anemo:
							x.AddStatMod(character.StatMod{
								Base:         modifier.NewBaseWithHitlag("scroll-4pc-nightsoul-anemo", 20*60),
								AffectedStat: attributes.AnemoP,
								Amount: func() ([]float64, bool) {
									nanemo[attributes.AnemoP] = dmgnightsoul
									return nanemo, true
								},
							})
						case attributes.Geo:
							x.AddStatMod(character.StatMod{
								Base:         modifier.NewBaseWithHitlag("scroll-4pc-nightsoul-geo", 20*60),
								AffectedStat: attributes.GeoP,
								Amount: func() ([]float64, bool) {
									ngeo[attributes.GeoP] = dmgnightsoul
									return ngeo, true
								},
							})
						default:
						}
						switch ele2 {
						case attributes.Electro:
							x.AddStatMod(character.StatMod{
								Base:         modifier.NewBaseWithHitlag("scroll-4pc-nightsoul-electro", 20*60),
								AffectedStat: attributes.ElectroP,
								Amount: func() ([]float64, bool) {
									nelectro[attributes.ElectroP] = dmgnightsoul
									return nelectro, true
								},
							})
						case attributes.Pyro:
							x.AddStatMod(character.StatMod{
								Base:         modifier.NewBaseWithHitlag("scroll-4pc-nightsoul-pyro", 20*60),
								AffectedStat: attributes.PyroP,
								Amount: func() ([]float64, bool) {
									npyro[attributes.PyroP] = dmgnightsoul
									return npyro, true
								},
							})
						case attributes.Cryo:
							x.AddStatMod(character.StatMod{
								Base:         modifier.NewBaseWithHitlag("scroll-4pc-nightsoul-cryo", 20*60),
								AffectedStat: attributes.CryoP,
								Amount: func() ([]float64, bool) {
									ncryo[attributes.CryoP] = dmgnightsoul
									return ncryo, true
								},
							})
						case attributes.Hydro:
							x.AddStatMod(character.StatMod{
								Base:         modifier.NewBaseWithHitlag("scroll-4pc-nightsoul-hydro", 20*60),
								AffectedStat: attributes.HydroP,
								Amount: func() ([]float64, bool) {
									nhydro[attributes.HydroP] = dmgnightsoul
									return nhydro, true
								},
							})
						case attributes.Dendro:
							x.AddStatMod(character.StatMod{
								Base:         modifier.NewBaseWithHitlag("scroll-4pc-nightsoul-dendro", 20*60),
								AffectedStat: attributes.DendroP,
								Amount: func() ([]float64, bool) {
									ndendro[attributes.DendroP] = dmgnightsoul
									return ndendro, true
								},
							})
						case attributes.Anemo:
							x.AddStatMod(character.StatMod{
								Base:         modifier.NewBaseWithHitlag("scroll-4pc-nightsoul-anemo", 20*60),
								AffectedStat: attributes.AnemoP,
								Amount: func() ([]float64, bool) {
									nanemo[attributes.AnemoP] = dmgnightsoul
									return nanemo, true
								},
							})
						case attributes.Geo:
							x.AddStatMod(character.StatMod{
								Base:         modifier.NewBaseWithHitlag("scroll-4pc-nightsoul-geo", 20*60),
								AffectedStat: attributes.GeoP,
								Amount: func() ([]float64, bool) {
									ngeo[attributes.GeoP] = dmgnightsoul
									return ngeo, true
								},
							})
						default:
						}
					}
					switch ele1 {
					case attributes.Electro:
						x.AddStatMod(character.StatMod{
							Base:         modifier.NewBaseWithHitlag("scroll-4pc-electro", 20*60),
							AffectedStat: attributes.ElectroP,
							Amount: func() ([]float64, bool) {
								electro[attributes.ElectroP] = dmg
								return electro, true
							},
						})
					case attributes.Pyro:
						x.AddStatMod(character.StatMod{
							Base:         modifier.NewBaseWithHitlag("scroll-4pc-pyro", 20*60),
							AffectedStat: attributes.PyroP,
							Amount: func() ([]float64, bool) {
								pyro[attributes.PyroP] = dmg
								return pyro, true
							},
						})
					case attributes.Cryo:
						x.AddStatMod(character.StatMod{
							Base:         modifier.NewBaseWithHitlag("scroll-4pc-cryo", 20*60),
							AffectedStat: attributes.CryoP,
							Amount: func() ([]float64, bool) {
								cryo[attributes.CryoP] = dmg
								return cryo, true
							},
						})
					case attributes.Hydro:
						x.AddStatMod(character.StatMod{
							Base:         modifier.NewBaseWithHitlag("scroll-4pc-hydro", 20*60),
							AffectedStat: attributes.HydroP,
							Amount: func() ([]float64, bool) {
								hydro[attributes.HydroP] = dmg
								return hydro, true
							},
						})
					case attributes.Dendro:
						x.AddStatMod(character.StatMod{
							Base:         modifier.NewBaseWithHitlag("scroll-4pc-dendro", 20*60),
							AffectedStat: attributes.DendroP,
							Amount: func() ([]float64, bool) {
								dendro[attributes.DendroP] = dmg
								return dendro, true
							},
						})
					case attributes.Anemo:
						x.AddStatMod(character.StatMod{
							Base:         modifier.NewBaseWithHitlag("scroll-4pc-anemo", 20*60),
							AffectedStat: attributes.AnemoP,
							Amount: func() ([]float64, bool) {
								anemo[attributes.AnemoP] = dmg
								return anemo, true
							},
						})
					case attributes.Geo:
						x.AddStatMod(character.StatMod{
							Base:         modifier.NewBaseWithHitlag("scroll-4pc-geo", 20*60),
							AffectedStat: attributes.GeoP,
							Amount: func() ([]float64, bool) {
								geo[attributes.GeoP] = dmg
								return geo, true
							},
						})
					default:
					}
					switch ele2 {
					case attributes.Electro:
						x.AddStatMod(character.StatMod{
							Base:         modifier.NewBaseWithHitlag("scroll-4pc-electro", 20*60),
							AffectedStat: attributes.ElectroP,
							Amount: func() ([]float64, bool) {
								electro[attributes.ElectroP] = dmg
								return electro, true
							},
						})
					case attributes.Pyro:
						x.AddStatMod(character.StatMod{
							Base:         modifier.NewBaseWithHitlag("scroll-4pc-pyro", 20*60),
							AffectedStat: attributes.PyroP,
							Amount: func() ([]float64, bool) {
								pyro[attributes.PyroP] = dmg
								return pyro, true
							},
						})
					case attributes.Cryo:
						x.AddStatMod(character.StatMod{
							Base:         modifier.NewBaseWithHitlag("scroll-4pc-cryo", 20*60),
							AffectedStat: attributes.CryoP,
							Amount: func() ([]float64, bool) {
								cryo[attributes.CryoP] = dmg
								return cryo, true
							},
						})
					case attributes.Hydro:
						x.AddStatMod(character.StatMod{
							Base:         modifier.NewBaseWithHitlag("scroll-4pc-hydro", 20*60),
							AffectedStat: attributes.HydroP,
							Amount: func() ([]float64, bool) {
								hydro[attributes.HydroP] = dmg
								return hydro, true
							},
						})
					case attributes.Dendro:
						x.AddStatMod(character.StatMod{
							Base:         modifier.NewBaseWithHitlag("scroll-4pc-dendro", 20*60),
							AffectedStat: attributes.DendroP,
							Amount: func() ([]float64, bool) {
								dendro[attributes.DendroP] = dmg
								return dendro, true
							},
						})
					case attributes.Anemo:
						x.AddStatMod(character.StatMod{
							Base:         modifier.NewBaseWithHitlag("scroll-4pc-anemo", 20*60),
							AffectedStat: attributes.AnemoP,
							Amount: func() ([]float64, bool) {
								anemo[attributes.AnemoP] = dmg
								return anemo, true
							},
						})
					case attributes.Geo:
						x.AddStatMod(character.StatMod{
							Base:         modifier.NewBaseWithHitlag("scroll-4pc-geo", 20*60),
							AffectedStat: attributes.GeoP,
							Amount: func() ([]float64, bool) {
								geo[attributes.GeoP] = dmg
								return geo, true
							},
						})
					default:
					}
				}
				c.Log.NewEvent("scroll 4pc proc'd", glog.LogWeaponEvent, char.Index).
					Write("trigger", key)
				return false
			}
		}

		switch char.Base.Element {
		case attributes.Anemo:
			c.Events.Subscribe(event.OnSwirlCryo, scrollfunc(attributes.Anemo, attributes.Cryo, "scroll-swirlcryo"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnSwirlElectro, scrollfunc(attributes.Anemo, attributes.Electro, "scroll-swirlelectro"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnSwirlHydro, scrollfunc(attributes.Anemo, attributes.Hydro, "scroll-swirlhydro"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnSwirlPyro, scrollfunc(attributes.Anemo, attributes.Pyro, "scroll-swirlpyro"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
		case attributes.Cryo:
			c.Events.Subscribe(event.OnSwirlCryo, scrollfunc(attributes.Anemo, attributes.Cryo, "scroll-swirlcryo"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnSuperconduct, scrollfunc(attributes.Electro, attributes.Cryo, "scroll-superconduct"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnCrystallizeCryo, scrollfunc(attributes.Geo, attributes.Cryo, "scroll-crystallizecryo"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnFrozen, scrollfunc(attributes.Hydro, attributes.Cryo, "scroll-frozen"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnMelt, scrollfunc(attributes.Pyro, attributes.Cryo, "scroll-melt"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
		case attributes.Dendro:
			c.Events.Subscribe(event.OnQuicken, scrollfunc(attributes.Dendro, attributes.Electro, "scroll-quicken"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnSpread, scrollfunc(attributes.Dendro, attributes.NoElement, "scroll-spread"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnBloom, scrollfunc(attributes.Dendro, attributes.Hydro, "scroll-bloom"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnBurning, scrollfunc(attributes.Dendro, attributes.Pyro, "scroll-burning"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
		case attributes.Electro:
			c.Events.Subscribe(event.OnSwirlElectro, scrollfunc(attributes.Anemo, attributes.Electro, "scroll-swirlelectro"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnSuperconduct, scrollfunc(attributes.Electro, attributes.Cryo, "scroll-superconduct"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnQuicken, scrollfunc(attributes.Dendro, attributes.Electro, "scroll-quicken"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnAggravate, scrollfunc(attributes.Electro, attributes.NoElement, "scroll-aggravate"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnHyperbloom, scrollfunc(attributes.Electro, attributes.NoElement, "scroll-hyperbloom"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnCrystallizeElectro, scrollfunc(attributes.Geo, attributes.Electro, "scroll-crystallizeelectro"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnElectroCharged, scrollfunc(attributes.Hydro, attributes.Electro, "scroll-electrocharged"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnOverload, scrollfunc(attributes.Pyro, attributes.Electro, "scroll-overload"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
		case attributes.Geo:
			c.Events.Subscribe(event.OnCrystallizeCryo, scrollfunc(attributes.Geo, attributes.Cryo, "scroll-crystallizecryo"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnCrystallizeElectro, scrollfunc(attributes.Geo, attributes.Electro, "scroll-crystallizeelectro"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnCrystallizeHydro, scrollfunc(attributes.Geo, attributes.Hydro, "scroll-crystallizehydro"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnCrystallizePyro, scrollfunc(attributes.Geo, attributes.Pyro, "scroll-crystallizepyro"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
		case attributes.Hydro:
			c.Events.Subscribe(event.OnSwirlHydro, scrollfunc(attributes.Anemo, attributes.Hydro, "scroll-swirlelhydro"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnFrozen, scrollfunc(attributes.Hydro, attributes.Cryo, "scroll-frozen"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnBloom, scrollfunc(attributes.Dendro, attributes.Hydro, "scroll-bloom"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnElectroCharged, scrollfunc(attributes.Hydro, attributes.Electro, "scroll-electrocharged"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnCrystallizeHydro, scrollfunc(attributes.Geo, attributes.Hydro, "scroll-crystallizehydro"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnVaporize, scrollfunc(attributes.Hydro, attributes.Pyro, "scroll-vaporize"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
		case attributes.Pyro:
			c.Events.Subscribe(event.OnSwirlPyro, scrollfunc(attributes.Anemo, attributes.Pyro, "scroll-swirlpyro"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnMelt, scrollfunc(attributes.Pyro, attributes.Cryo, "scroll-melt"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnBurning, scrollfunc(attributes.Dendro, attributes.Pyro, "scroll-burning"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnBurgeon, scrollfunc(attributes.Pyro, attributes.NoElement, "scroll-burgeon"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnOverload, scrollfunc(attributes.Pyro, attributes.Electro, "scroll-overload"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnCrystallizePyro, scrollfunc(attributes.Geo, attributes.Pyro, "scroll-crystallizepyro"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
			c.Events.Subscribe(event.OnVaporize, scrollfunc(attributes.Hydro, attributes.Pyro, "scroll-vaporize"), fmt.Sprintf("scroll-4pc-%v", char.Base.Key.String()))
		default:
		}
	}
	return &s, nil
}
