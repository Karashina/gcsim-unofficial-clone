package klee

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

const (
	a1IcdKey   = "a1-icd"
	a1SparkKey = "a1-spark"
)

// Jumpy Dumptyと通常攻撃がダメージを与えると、クレーは50%の確率で爆裂火花を獲得する。
// この爆裂火花は次の重撃で消費され、スタミナ消費なしでダメージが50%増加する。
func (c *char) makeA1CB() combat.AttackCBFunc {
	if c.Base.Ascension < 1 {
		return nil
	}
	return func(a combat.AttackCB) {
		if c.StatusIsActive(a1IcdKey) {
			return
		}
		if c.Core.Rand.Float64() < 0.5 {
			return
		}
		c.AddStatus(a1IcdKey, 60*5, true)

		if !c.StatusIsActive(a1SparkKey) {
			c.AddStatus(a1SparkKey, 60*30, true)
		}
	}
}

const a4ICDKey = "klee-a4-icd"

// クレーの重撃が会心した場合、パーティ全員が元素エネルギーを2回復する。
func (c *char) makeA4CB() combat.AttackCBFunc {
	if c.Base.Ascension < 4 {
		return nil
	}
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if !a.IsCrit {
			return
		}
		if c.StatusIsActive(a4ICDKey) {
			return
		}
		c.AddStatus(a4ICDKey, 0.6*60, true)
		for _, x := range c.Core.Player.Chars() {
			x.AddEnergy("klee-a4", 2)
		}
	}
}
