package nefer

import (
	_ "embed"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
)

// Special Lunar Bloom damage handler for Nefer
func (c *char) onLunarBloomNeferSpecial(args ...interface{}) bool {
	n := args[0].(combat.Target)
	ae := args[1].(*combat.AttackEvent)

	if ae.Info.AttackTag != attacks.AttackTagLBDamage {
		return false
	}
	c.Core.Events.Emit(event.OnLunarBloom, n, ae)
	return false
}

// Register Nefer's special Lunar Charged callback
func (c *char) InitLCallback() {
	c.Core.Events.Subscribe(event.OnEnemyHit, c.onLunarBloomNeferSpecial, "lc-nefer-special")
}
