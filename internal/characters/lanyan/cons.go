package lanyan

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/core/player/shield"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const c2IcdKey = "lanyan-c2-icd"

func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)

		// ignore if c2 on icd
		if c.StatusIsActive(c2IcdKey) {
			return false
		}
		// On normal attack
		if ae.Info.AttackTag != attacks.AttackTagNormal {
			return false
		}
		// make sure the person triggering the attack is on field still
		if ae.Info.ActorIndex != c.Core.Player.Active() {
			return false
		}
		c.AddStatus(c2IcdKey, 2*60, true)

		existingShield := c.Core.Player.Shields.Get(shield.LanyanSkill)
		amt := 0.0
		if existingShield != nil {
			amt = min(c.shieldamt, existingShield.CurrentHP()+c.shieldamt*0.4)
		}
		// regenerate shield
		c.Core.Player.Shields.Replace(&shield.Tmpl{
			ActorIndex: c.Index,
			Target:     -1,
			Src:        c.shieldsrc,
			ShieldType: shield.LanyanSkill,
			Name:       "Swallow-Wisp Shield",
			HP:         amt,
			Ele:        c.shieldele,
			Expires:    c.shieldexp,
		})
		return false
	}, "lanyan-c2")
}

func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.EM] = 60
	for _, char := range c.Core.Player.Chars() {
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("lanyan-c4", 12*60),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}
}
