package ineffa

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/shield"
)

// Generate shield for the character
func (c *char) genShield(src string, shieldamt float64) {
	c.c1()
	c.Core.Tasks.Add(func() {
		c.Core.Player.Shields.Add(&shield.Tmpl{
			ActorIndex: c.Index,
			Target:     -1,
			Src:        c.Core.F,
			ShieldType: shield.IneffaSkill,
			Name:       src,
			HP:         shieldamt,
			Ele:        attributes.Hydro,
			Expires:    c.Core.F + 20*60,
		})
	}, 1)
}

// Calculate shield HP based on skill level and stats
func (c *char) shieldHP() float64 {
	return shieldPct[c.TalentLvlSkill()]*c.TotalAtk() + shieldCst[c.TalentLvlSkill()]
}
