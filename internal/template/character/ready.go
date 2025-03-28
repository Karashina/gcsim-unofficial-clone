package character

import (
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/glog"
)

func (c *Character) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	// for dash and charge need to check for stam usage as well

	switch a {
	case action.ActionBurst:
		if !c.Core.Flags.IgnoreBurstEnergy && c.Energy != c.EnergyMax {
			return false, action.InsufficientEnergy
		}
		if c.AvailableCDCharge[a] <= 0 {
			return false, action.BurstCD
		}
	case action.ActionSkill:
		if c.AvailableCDCharge[a] <= 0 {
			return false, action.SkillCD
		}
	case action.ActionCharge:
		req := c.Core.Player.AbilStamCost(c.Index, a, p)
		if c.Core.Player.Stam < req {
			c.Core.Log.NewEvent("insufficient stam: charge attack", glog.LogWarnings, -1).
				Write("have", c.Core.Player.Stam)
			return false, action.InsufficientStamina
		}
	case action.ActionDash:
		req := c.Core.Player.AbilStamCost(c.Index, a, p)
		if c.Core.Player.Stam < req {
			c.Core.Log.NewEvent("insufficient stam: dash", glog.LogWarnings, -1).
				Write("have", c.Core.Player.Stam)
			return false, action.InsufficientStamina
		}
		if c.Core.Player.Active() == c.Index && c.Core.Player.DashLockout && c.Core.Player.DashCDExpirationFrame > c.Core.F {
			c.Core.Log.NewEvent("dash on cooldown", glog.LogWarnings, -1).
				Write("dash_cd_expiration", c.Core.Player.DashCDExpirationFrame-c.Core.F)
			return false, action.DashCD
		}
		if c.Core.Player.Active() != c.Index && c.DashLockout && c.RemainingDashCD > 0 {
			return false, action.DashCD
		}
	}
	return true, action.NoFailure
}
