package raiden

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 祕伝一刀知成状態終了時、周囲の全パーティメンバー（雷電将軍除く）の攻撃力+30%、10秒間。
func (c *char) c4() {
	m := make([]float64, attributes.EndStatType)
	m[attributes.ATKP] = 0.3

	for i, char := range c.Core.Player.Chars() {
		if i == c.Index {
			continue
		}
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("raiden-c4", 600),
			AffectedStat: attributes.ATKP,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}
}

func (c *char) c6(ac combat.AttackCB) {
	if c.Base.Cons < 6 {
		return
	}
	if c.Core.F < c.c6ICD {
		return
	}
	if c.c6Count == 5 {
		return
	}
	c.c6ICD = c.Core.F + 60
	c.c6Count++
	c.Core.Log.NewEvent("raiden c6 triggered", glog.LogCharacterEvent, c.Index).
		Write("next_icd", c.c6ICD).
		Write("count", c.c6Count)
	for i, char := range c.Core.Player.Chars() {
		if i == c.Index {
			continue
		}
		char.ReduceActionCooldown(action.ActionBurst, 60)
	}
}
