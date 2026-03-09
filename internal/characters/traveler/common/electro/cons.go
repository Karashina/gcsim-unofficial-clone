package electro

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 2凸 - Violet Vehemence
// 轟く雷で生じた落雷が敵に命中すると、8秒間雷元素耐性を15%減少させる。
func (c *Traveler) c2() combat.AttackCBFunc {
	if c.Base.Cons < 2 {
		return nil
	}
	return func(a combat.AttackCB) {
		e, ok := a.Target.(*enemy.Enemy)
		if !ok {
			return
		}
		e.AddResistMod(combat.ResistMod{
			Base:  modifier.NewBaseWithHitlag("travelerelectro-c2", 480),
			Ele:   attributes.Electro,
			Value: -0.15,
		})
	}
}

// 4凸 - 雷影剣で生成されたアバンダンスアミュレットを獲得した時、キャラクターのエネルギーが
// 35%未満の場合、アミュレットのエネルギー回復量が100%増加。
func (c *Traveler) c4(buffEnergy float64) float64 {
	if c.Base.Cons >= 4 {
		collector := c.Core.Player.ActiveChar()
		currentEnergyP := collector.Energy / collector.EnergyMax
		if currentEnergyP < 0.35 {
			buffEnergy *= 2
		}
	}
	return buffEnergy
}

// World-Shaker
// 轟く雷でトリガーされた落雷2回ごとにダメージが大幅に増加し、
// 次の落雷は元のダメージの200%を与える [..]
// * 雷旅人の6凸は乗算バフ
func (c *Traveler) c6Damage(ai *combat.AttackEvent) {
	if c.Base.Cons >= 6 {
		c.burstC6Hits++
		if c.burstC6Hits >= 3 {
			// TODO これで正しく乗算される？代わりにmodを使うべき？
			ai.Info.Mult *= 2
			c.burstC6Hits = 0
			c.burstC6WillGiveEnergy = true
		}
	}
}

// World-Shaker
//
//	[..] かつ現在のキャラクターに追加で1エネルギーを回復する。
func (c *Traveler) c6Energy() combat.AttackCBFunc {
	return func(_ combat.AttackCB) {
		if c.burstC6WillGiveEnergy {
			active := c.Core.Player.ActiveChar()
			active.AddEnergy("travelerelectro-c6", 1)
			c.burstC6WillGiveEnergy = false
		}
	}
}
