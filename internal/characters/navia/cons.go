package navia

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const c2IcdKey = "navia-c2-icd"

// セレモニアルクリスタルショット使用時に消費したCrystal Shrapnelのスタック毎に
// ナヴィアのエネルギーを3回復し、As the Sunlit Sky's Singing SaluteのCDを1秒短縮。
// エネルギーは最大9回復、CDは最大3秒短縮可能。
func (c *char) c1(shrapnel int) {
	if c.Base.Cons < 1 {
		return
	}
	count := min(shrapnel, 3)
	c.ReduceActionCooldown(action.ActionBurst, count*60)
	c.AddEnergy("navia-c1-energy", float64(count*3))
}

// 消費したCrystal Shrapnelのスタック毎に、セレモニアルクリスタルショットの
// 会心率が12%増加。会心率は最大36%まで増加可能。
// また、セレモニアルクリスタルショットが敵に命中すると、
// As the Sunlit Sky's Singing Saluteの砲撃支援1発が命中地点付近に着弾。
// セレモニアルクリスタルショット使用毎に最大1回発動可能。
// この砲撃支援のダメージは元素爆発ダメージ扱い。
func (c *char) c2() combat.AttackCBFunc {
	if c.Base.Cons < 2 {
		return nil
	}
	return func(a combat.AttackCB) {
		e := a.Target.(*enemy.Enemy)
		if e.Type() != targets.TargettableEnemy {
			return
		}
		if c.StatusIsActive(c2IcdKey) {
			return
		}
		c.AddStatus(c2IcdKey, 0.25*60, true)
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "The President's Pursuit of Victory",
			AttackTag:  attacks.AttackTagElementalBurst,
			ICDTag:     attacks.ICDTagElementalBurst,
			ICDGroup:   attacks.ICDGroupNaviaBurst,
			StrikeType: attacks.StrikeTypeBlunt,
			PoiseDMG:   50,
			Element:    attributes.Geo,
			Durability: 25,
			Mult:       burst[1][c.TalentLvlBurst()],
		}
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(geometry.CalcRandomPointFromCenter(e.Pos(), 0, 1.2, c.Core.Rand), nil, 3),
			0,
			30, // somewhere between 28-31
			c.burstCB(),
			c.c4(),
		)
	}
}

// As the Sunlit Sky's Singing Saluteが敵に命中すると、
// その敵の岩元素耐性が8秒間20%低下する。
func (c *char) c4() combat.AttackCBFunc {
	if c.Base.Cons < 4 {
		return nil
	}
	return func(a combat.AttackCB) {
		e := a.Target.(*enemy.Enemy)
		if e.Type() != targets.TargettableEnemy {
			return
		}
		e.AddResistMod(combat.ResistMod{
			Base:  modifier.NewBaseWithHitlag("navia-c4-shred", 8*60),
			Ele:   attributes.Geo,
			Value: -0.2,
		})
	}
}
