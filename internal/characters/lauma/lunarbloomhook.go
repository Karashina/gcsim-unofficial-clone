package lauma

import (
	_ "embed"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
)

// Lauma専用のLunar Bloomダメージハンドラー
func (c *char) onLunarBloomLaumaSpecial(args ...interface{}) bool {
	ae := args[1].(*combat.AttackEvent)

	if ae.Info.AttackTag != attacks.AttackTagLBDamage {
		return false
	}
	return false
}

// Lauma専用のLunar Chargedコールバックを登録
func (c *char) InitLCallback() {
	c.Core.Events.Subscribe(event.OnEnemyHit, c.onLunarBloomLaumaSpecial, "lc-lauma-special")
}
