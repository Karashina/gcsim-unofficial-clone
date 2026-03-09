package tartaglia

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
)

// 断流の持続時間を8秒延長する。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	c.riptideDuration += 8 * 60
}

// タルタリヤが断流・近接スタンス中に会心が発生した場合、
// 通常攻撃と重撃が敵に断流を付与する。
func (c *char) makeA4CB() combat.AttackCBFunc {
	if c.Base.Ascension < 4 {
		return nil
	}
	return func(a combat.AttackCB) {
		if a.IsCrit {
			t, ok := a.Target.(*enemy.Enemy)
			if !ok {
				return
			}
			c.applyRiptide("melee", t)
		}
	}
}
