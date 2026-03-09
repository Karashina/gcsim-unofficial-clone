package razor

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 爪雷のCDが18%減少する。
func (c *char) a1CDReduction(cd int) int {
	if c.Base.Ascension < 1 {
		return cd
	}
	return int(float64(cd) * 0.82)
}

// 雷牙使用時に爪雷のCDをリセットする。
func (c *char) a1CDReset() {
	if c.Base.Ascension < 1 {
		return
	}
	c.ResetActionCooldown(action.ActionSkill)
}

// Razorのエネルギーが50%未満の時、元素チャージ効率が30%上昇する。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	c.a4Bonus = make([]float64, attributes.EndStatType)
	c.a4Bonus[attributes.ER] = 0.3
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("razor-a4", -1),
		AffectedStat: attributes.ER,
		Amount: func() ([]float64, bool) {
			if c.Energy/c.EnergyMax >= 0.5 {
				return nil, false
			}
			return c.a4Bonus, true
		},
	})
}
