package shenhe

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const c4BuffKey = "shenhe-c4"

func (c *char) c2(active *character.CharWrapper, dur int) {
	active.AddAttackMod(character.AttackMod{
		Base: modifier.NewBaseWithHitlag("shenhe-c2", dur),
		Amount: func(ae *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
			if ae.Info.Element != attributes.Cryo {
				return nil, false
			}
			return c.c2buff, true
		},
	})
}

// 氷羽の効果を受けたキャラクターがダメージボーナスを発動した時、申鶴は霜霊詣スタックを獲得する:
//
//   - 申鶴が仕組みの春神を使用時、全ての霜霊詣スタックを消費し、
//     消費したスタック1つにつきその仕組みの春神のダメージが5%増加する。
//
// 最奇50スタック。スタックは60秒間持続。
func (c *char) c4() float64 {
	if c.Base.Cons < 4 {
		return 0
	}
	if !c.StatusIsActive(c4BuffKey) {
		c.c4count = 0
		return 0
	}
	dmgBonus := 0.05 * float64(c.c4count)
	c.Core.Log.NewEvent("shenhe-c4 adding dmg bonus", glog.LogCharacterEvent, c.Index).
		Write("stacks", c.c4count).
		Write("dmg_bonus", dmgBonus)
	c.c4count = 0
	c.DeleteStatus(c4BuffKey)
	return dmgBonus
}

// C4スタックはダメージが与えられた後に獲得される（前ではなく）
// https://library.keqingmains.com/evidence/characters/cryo/shenhe?q=shenhe#c4-insight
func (c *char) c4CB(a combat.AttackCB) {
	// 全て期限切れの場合スタックをゼロにリセット
	if !c.StatusIsActive(c4BuffKey) {
		c.c4count = 0
	}
	if c.c4count < 50 {
		c.c4count++
		c.Core.Log.NewEvent("shenhe-c4 stack gained", glog.LogCharacterEvent, c.Index).
			Write("stacks", c.c4count)
	}
	c.AddStatus(c4BuffKey, 3600, true) // 60 s
}
