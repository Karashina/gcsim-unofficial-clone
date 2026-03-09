package wriothesley

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
)

func (c *char) Dash(p map[string]int) (action.Info, error) {
	// 通常攻撃/スキル以外 -> スキルはsavedNormalCounterをリセットすべき
	switch c.Core.Player.LastAction.Type {
	case action.ActionAttack:
	case action.ActionSkill:
	default:
		c.savedNormalCounter = 0
	}

	return c.Character.Dash(p)
}
