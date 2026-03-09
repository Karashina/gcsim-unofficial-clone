package geo

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
)

// 星落としの剣のCTを2秒短縮する。
func (c *Traveler) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	c.skillCD -= 2 * 60
}

// 通常攻撃コンボの最終段が崩落をトリガーし、攻撃力の60%の岩元素範囲ダメージを与える。
func (c *Traveler) a4() {
	if c.Base.Ascension < 4 || c.NormalCounter != c.NormalHitNum-1 {
		return
	}
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Frenzied Rockslide (A4)",
		AttackTag:  attacks.AttackTagNormal,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		PoiseDMG:   13.5,
		Element:    attributes.Geo,
		Durability: 25,
		Mult:       0.6,
	}
	c.QueueCharTask(func() {
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 1.2}, 2.4),
			0,
			0,
		)
	}, a4Hitmark[c.gender])
}
