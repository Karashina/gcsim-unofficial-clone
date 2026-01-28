package lauma

import (
	_ "embed"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
)

// Special Lunar Bloom damage handler for Lauma
func (c *char) onLunarBloomLaumaSpecial(args ...interface{}) bool {
	ae := args[1].(*combat.AttackEvent)

	if ae.Info.AttackTag != attacks.AttackTagLBDamage {
		return false
	}
	return false
}

// Register Ineffa's special Lunar Charged callback
func (c *char) InitLCallback() {
	c.Core.Events.Subscribe(event.OnEnemyHit, c.onLunarBloomLaumaSpecial, "lc-lauma-special")
}
