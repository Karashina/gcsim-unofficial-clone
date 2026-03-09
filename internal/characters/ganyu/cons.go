package ganyu

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	c1Key = "ganyu-c1"
	c4Key = "ganyu-c4"
	c4Dur = 180
	c6Key = "ganyu-c6"
)

// 甘雨C1: チャージレベル2の霄花矢または霄花矢・霄花のダメージを受けた敵の氷元素耐性-15%、6秒間。
// 命中時、甘雨の元素エネルギーを2回復。この効果はチャージレベル2の霄花矢ごとに1回のみ発動。
func (c *char) c1() combat.AttackCBFunc {
	if c.Base.Cons < 1 {
		return nil
	}
	done := false

	return func(a combat.AttackCB) {
		e := a.Target.(*enemy.Enemy)
		if e.Type() != targets.TargettableEnemy {
			return
		}
		e.AddResistMod(combat.ResistMod{
			Base:  modifier.NewBaseWithHitlag(c1Key, 300),
			Ele:   attributes.Cryo,
			Value: -0.15,
		})
		if done {
			return
		}
		done = true
		c.AddEnergy(c1Key, 2)
	}
}

func (c *char) c4() {
	m := make([]float64, attributes.EndStatType)
	for _, char := range c.Core.Player.Chars() {
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase(c4Key, -1),
			Amount: func(_ *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				x, ok := t.(*enemy.Enemy)
				if !ok {
					return nil, false
				}
				// 期限切れ時にスタックをリセット
				if !x.StatusIsActive(c4Key) {
					x.RemoveTag(c4Key)
					return nil, false
				}
				m[attributes.DmgP] = float64(x.GetTag(c4Key)) * 0.05
				return m, true
			},
		})
	}
}
