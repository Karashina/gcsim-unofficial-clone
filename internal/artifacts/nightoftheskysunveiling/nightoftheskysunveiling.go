package nightoftheskysunveiling

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
	core.RegisterSetFunc(keys.NightOfTheSkysUnveiling, NewSet)
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
		m[attributes.EM] = 80
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("notsu-2pc", -1),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}
	if count >= 4 {

		notsucb := func(args ...interface{}) bool {
			if c.Player.ActiveChar().Index != char.Index {
				return false
			}
			char.AddStatus("gleamingmoon-key-notsu", 4*60, true)
			m := make([]float64, attributes.EndStatType)
			char.AddStatMod(character.StatMod{
				Base:         modifier.NewBase("notsu-4pc-cr", 4*60),
				AffectedStat: attributes.CR,
				Amount: func() ([]float64, bool) {
					m[attributes.CR] = 0.0
					if char.MoonsignAscendant {
						m[attributes.CR] = 0.30
					} else if char.MoonsignNascent {
						m[attributes.CR] = 0.15
					}
					return m, true
				},
			})
			return false
		}
		c.Events.Subscribe(event.OnLunarBloom, notsucb, "notsu-4pc-lb")
		c.Events.Subscribe(event.OnLunarCharged, notsucb, "notsu-4pc-lc")

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
