package ayaka

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	c1ICDKey = "ayaka-c1-icd"
)

// 綾華1凸のコールバック（通常攻撃/重撃にアタッチ）
// 神里綾華の通常攻撃または重撃が氷元素ダメージを与えた場合　50%の確率で
// 神里流・氷華のCDを0.3秒短縮する。
// この効果は0.1秒ごとに1回発動する。
func (c *char) c1(a combat.AttackCB) {
	if c.Base.Cons < 1 {
		return
	}
	if a.AttackEvent.Info.Element != attributes.Cryo {
		return
	}
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.StatusIsActive(c1ICDKey) {
		return
	}
	if c.Core.Rand.Float64() < .5 {
		return
	}
	c.ReduceActionCooldown(action.ActionSkill, 18)
	c.AddStatus(c1ICDKey, 6, true)
}

// 綾華4凸のコールバック（元素爆発ヒットにアタッチ）
// 神里流・霜滅の霜華石にダメージを受けた敵の防御力が6秒間30%減少する。
func (c *char) c4(a combat.AttackCB) {
	if c.Base.Cons < 4 {
		return
	}
	if a.Damage == 0 {
		return
	}

	e, ok := a.Target.(*enemy.Enemy)
	if !ok {
		return
	}
	e.AddDefMod(combat.DefMod{
		Base:  modifier.NewBaseWithHitlag("ayaka-c4", 60*6),
		Value: -0.3,
	})
}

// 重撃ヒットにアタッチされる綾華6凸のコールバック
func (c *char) c6(a combat.AttackCB) {
	if c.Base.Cons < 6 {
		return
	}
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}

	if !c.c6CDTimerAvail {
		return
	}
	c.c6CDTimerAvail = false

	c.QueueCharTask(func() {
		c.DeleteAttackMod("ayaka-c6")
		c.QueueCharTask(func() {
			c.c6CDTimerAvail = true
			c.c6AddBuff()
		}, 600)
	}, 30)
}

func (c *char) c6AddBuff() {
	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = 2.98

	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("ayaka-c6", -1),
		Amount: func(atk *combat.AttackEvent, _ combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagExtra {
				return nil, false
			}
			return m, true
		},
	})
}
