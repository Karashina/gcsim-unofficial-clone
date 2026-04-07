package varka

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func (c *char) c4Init() {
	swirlMap := map[event.Event]attributes.Stat{
		event.OnSwirlHydro:   attributes.HydroP,
		event.OnSwirlPyro:    attributes.PyroP,
		event.OnSwirlCryo:    attributes.CryoP,
		event.OnSwirlElectro: attributes.ElectroP,
	}

	for ev, eleStat := range swirlMap {
		eleStat := eleStat // capture for closure
		c.Core.Events.Subscribe(ev, func(args ...interface{}) bool {
			atk := args[1].(*combat.AttackEvent)
			// Only triggers when Varka triggers the Swirl
			if atk.Info.ActorIndex != c.Index {
				return false
			}

			// 全パーティメンバーに適用
			for _, char := range c.Core.Player.Chars() {
				m := make([]float64, attributes.EndStatType)
				m[attributes.AnemoP] = 0.20
				m[eleStat] = 0.20

				char.AddStatMod(character.StatMod{
					Base:         modifier.NewBaseWithHitlag(c4Key, 10*60),
					AffectedStat: attributes.NoStat,
					Amount: func() ([]float64, bool) {
						return m, true
					},
				})
			}

			return false
		}, fmt.Sprintf("varka-c4-%v", ev))
	}
}
