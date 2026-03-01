package varka

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// C1 - "Come, Friend, Let Us Dance Beneath the Moon's Soft Glow"
// +1 FWA charge and Lyrical Libation effect
// Handled in Init (fwaMaxCharges) and enterSturmUndDrang (c1LyricalKey status)
// The 200% DMG mult is handled in skill.go fourWindsAscension and charge.go azureDevour

// C2 - "When Dawn Breaks, Our Journey Shall Take Flight"
// Additional Anemo strike equal to 800% ATK on FWA or Azure Devour
// Handled in skill.go c2Strike and c2Init

// C4 - "For None May Take From Us Our Freedom of Song"
// When Varka triggers a Swirl reaction, all nearby party members gain
// 20% Anemo DMG Bonus and the corresponding Elemental DMG Bonus for 10s.
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

			// Apply to all party members
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

// C6 - "Beloved Mondstadt, Steadfast You Shall Shine"
// FWA→Azure chain, Azure→FWA chain without charge consumption
// The window statuses (c6FWAWindowKey, c6AzureWindowKey) are set in:
//   - skill.go fourWindsAscension() → sets c6FWAWindowKey
//   - charge.go azureDevour() → sets c6AzureWindowKey
// and consumed in fourWindsAscension and azureDevour respectively
//
// C6 CRIT DMG from A4 stacks is handled in asc.go a4Apply()
