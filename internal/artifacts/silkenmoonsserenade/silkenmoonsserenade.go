package silkenmoonsserenade

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterSetFunc(keys.SilkenMoonsSerenade, NewSet)
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

	if count >= 2 {
		m := make([]float64, attributes.EndStatType)
		m[attributes.ER] = 0.20
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("sms-2pc", -1),
			AffectedStat: attributes.ER,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}
	if count >= 4 {

		emval := 0.0
		nascent := false
		ascendant := false
		for _, char := range c.Player.Chars() {
			if char.MoonsignAscendant {
				ascendant = true
				nascent = false
				break
			}
			if char.MoonsignNascent {
				nascent = true
			}
		}
		if ascendant {
			emval = 120
		} else if nascent {
			emval = 60
		}

		c.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
			atk := args[1].(*combat.AttackEvent)
			if atk.Info.ActorIndex != char.Index {
				return false
			}
			if atk.Info.Element == attributes.Physical {
				return false
			}
			char.AddStatus("gleamingmoon-key-sms", 8*60, true)
			m := make([]float64, attributes.EndStatType)
			m[attributes.EM] = emval
			for _, ch := range c.Player.Chars() {
				ch.AddStatMod(character.StatMod{
					Base:         modifier.NewBase("sms-4pc-em", 8*60),
					AffectedStat: attributes.EM,
					Amount: func() ([]float64, bool) {
						return m, true
					},
				})
			}
			return false
		}, "sms-4pc")

		for _, ch := range c.Player.Chars() {
			ch.AddLBReactBonusMod(character.LBReactBonusMod{
				Base: modifier.NewBase("gleamingmoon-lb", -1),
				Amount: func(ai combat.AttackInfo) (float64, bool) {
					count := 0
					sms := false
					notsu := false
					for _, chr := range c.Player.Chars() {
						if chr.StatusIsActive("gleamingmoon-key-sms") || !sms {
							count++
							sms = true
						}
						if chr.StatusIsActive("gleamingmoon-key-notsu") || !notsu {
							count++
							notsu = true
						}
					}
					val := 0.1 * float64(count)
					return val, false
				},
			})
			ch.AddLCReactBonusMod(character.LCReactBonusMod{
				Base: modifier.NewBase("gleamingmoon-lc", -1),
				Amount: func(ai combat.AttackInfo) (float64, bool) {
					count := 0
					sms := false
					notsu := false
					for _, chr := range c.Player.Chars() {
						if chr.StatusIsActive("gleamingmoon-key-sms") || !sms {
							count++
							sms = true
						}
						if chr.StatusIsActive("gleamingmoon-key-notsu") || !notsu {
							count++
							notsu = true
						}
					}
					val := 0.1 * float64(count)
					return val, false
				},
			})
		}
	}

	return &s, nil
}

