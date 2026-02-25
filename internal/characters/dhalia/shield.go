package dhalia

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/shield"
)

func (c *char) genShield(src string, shieldamt float64) {
	existingShield := c.Core.Player.Shields.Get(shield.DhaliaBurst)
	if existingShield != nil {
		shieldamt += existingShield.CurrentHP()
	}

	c.shieldExpiry = c.Core.F + c.burstdur

	// add shield
	c.Core.Tasks.Add(func() {
		c.Core.Player.Shields.Add(&shield.Tmpl{
			ActorIndex: c.Index,
			Target:     -1,
			Src:        c.Core.F,
			ShieldType: shield.DhaliaBurst,
			Name:       src,
			HP:         shieldamt,
			Ele:        attributes.Hydro,
			Expires:    c.shieldExpiry,
		})
	}, 1)
}

func (c *char) shieldHP() float64 {
	return shieldPct[c.TalentLvlBurst()]*c.MaxHP() + shieldCst[c.TalentLvlBurst()]
}

func (c *char) regenShirld() {
	c.Core.Events.Subscribe(event.OnShieldBreak, func(args ...interface{}) bool {
		shd := args[0].(*shield.Tmpl)
		if shd.ShieldType != shield.DhaliaBurst {
			return false
		}
		if !c.StatusIsActive(burstkey) {
			return false
		}
		if c.favoniusfavor <= 0 {
			return false
		}
		c.favoniusfavor--
		// add shield
		c.Core.Tasks.Add(func() {
			c.Core.Player.Shields.Add(&shield.Tmpl{
				ActorIndex: c.Index,
				Target:     -1,
				Src:        c.Core.F,
				ShieldType: shield.DhaliaBurst,
				Name:       "dhalia-shield-regen",
				HP:         c.shieldHP(),
				Ele:        attributes.Hydro,
				Expires:    c.shieldExpiry,
			})
		}, 1)
		if c.Base.Cons >= 2 {
			c.Core.Player.Shields.AddShieldBonusMod("dhalia-c2", 12*60, func() (float64, bool) {
				if c.Tags["shielded"] == 0 {
					return 0, false
				}
				return 0.25, true
			})
		}
		return false
	}, "dhalia-shield-regen")
}
