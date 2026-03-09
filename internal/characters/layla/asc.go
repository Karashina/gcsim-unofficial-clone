package layla

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/shield"
)

// 安眠の帖がアクティブの間、布が星を１つ獲得するたびに「安眠」効果が発動:
//
// - 安眠の帖の影響を受けたキャラクターのシールド強度が6%増加する。
//
// - この効果は最大4スタックまで重複でき、安眠の帖が消えるまで持続する。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	c.Core.Player.Shields.AddShieldBonusMod("layla-a1", -1, func() (float64, bool) {
		if exist := c.Core.Player.Shields.Get(shield.LaylaSkill); exist == nil {
			return 0, false
		}
		return float64(c.a1Stack) * 0.06, false
	})
}

// ひとしきりの夜のシューティングスターのダメージがレイラの最大HPの1.5%分増加する。
func (c *char) a4() float64 {
	if c.Base.Ascension < 4 {
		return 0
	}
	return 0.015 * c.MaxHP()
}
