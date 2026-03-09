package jean

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

// ジンの通常攻撃がヒットすると　50%の確率でチーム全員のHPをジンの攻撃力の15%分回復する。
func (c *char) makeA1CB() combat.AttackCBFunc {
	if c.Base.Ascension < 1 {
		return nil
	}
	done := false
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if done {
			return
		}
		done = true

		snap := a.AttackEvent.Snapshot
		if c.Core.Rand.Float64() < 0.5 {
			c.Core.Player.Heal(info.HealInfo{
				Caller:  c.Index,
				Target:  -1,
				Message: "Wind Companion",
				Src:     snap.Stats.TotalATK() * .15,
				Bonus:   c.Stat(attributes.Heal),
			})
		}
	}
}

// 蒲公英の風を使用するとエネルギーの20%を回復する。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	c.AddEnergy("jean-a4", 16)
}
