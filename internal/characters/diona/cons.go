package diona

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func (c *char) c2() {
	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = .15
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("diona-c2", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			return m, atk.Info.AttackTag == attacks.AttackTagElementalArt
		},
	})
}

func (c *char) c6() {
	// 6凸は元素爆発の持続中有効
	// 12.5秒持続、0.5秒ごとにtick; アクティブキャラに2秒間のバフを付与
	for i := 30; i <= 750; i += 30 {
		c.Core.Tasks.Add(func() {
			if !c.Core.Combat.Player().IsWithinArea(c.burstBuffArea) {
				return
			}
			// アクティブキャラに元素熟知200を追加
			active := c.Core.Player.ActiveChar()
			if active.CurrentHPRatio() > 0.5 {
				active.AddStatMod(character.StatMod{
					Base:         modifier.NewBaseWithHitlag("diona-c6", 120),
					AffectedStat: attributes.EM,
					Amount: func() ([]float64, bool) {
						return c.c6buff, true
					},
				})
			} else {
				// HP50%以下の場合、受ける治療効果アップを追加
				// ボーナスは120フレームのみ持続
				active.AddHealBonusMod(character.HealBonusMod{
					Base: modifier.NewBaseWithHitlag("diona-c6-healbonus", 120),
					Amount: func() (float64, bool) {
						// このログは必要か？
						c.Core.Log.NewEvent("diona c6 incomming heal bonus activated", glog.LogCharacterEvent, c.Index)
						return 0.3, false
					},
				})
				c.Tags["c6bonus-"+active.Base.Key.String()] = c.Core.F + 120
			}
		}, i)
	}
}
