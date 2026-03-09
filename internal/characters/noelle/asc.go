package noelle

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/shield"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

const a1IcdKey = "noelle-a1-icd"

// ノエルがパーティーにいるがフィールド上にいない時、
// アクティブキャラクターのHPが30%を下回ると自動発動:
// ノエルの防御力400%分のダメージを吸収するシールドを生成。持続20秒。
// 全元素・物理ダメージに対して150%の吸収効率。
// この効果は60秒に1回のみ発動。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	c.Core.Events.Subscribe(event.OnPlayerHPDrain, func(args ...interface{}) bool {
		di := args[0].(*info.DrainInfo)
		if di.Amount <= 0 {
			return false
		}
		if c.StatusIsActive(a1IcdKey) {
			return false
		}
		active := c.Core.Player.ActiveChar()
		if active.CurrentHPRatio() >= 0.3 {
			return false
		}
		c.AddStatus(a1IcdKey, 3600, false)
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "A1 Shield",
			AttackTag:  attacks.AttackTagNone,
		}
		snap := c.Snapshot(&ai)

		// シールドを追加
		c.Core.Player.Shields.Add(&shield.Tmpl{
			ActorIndex: c.Index,
			Target:     active.Index,
			Src:        c.Core.F,
			ShieldType: shield.NoelleA1,
			Name:       "Noelle A1",
			HP:         snap.Stats.TotalDEF() * 4,
			Ele:        attributes.Cryo,
			Expires:    c.Core.F + 1200, // 20秒
		})
		return false
	}, "noelle-a1")
}

// ノエルの通常攻撃または重撃が敵に4回命中するごとに、
// 護心のCDが1秒短縮される。
// ヒットは0.1秒ごとに1回カウントされる。
func (c *char) makeA4CB() combat.AttackCBFunc {
	if c.Base.Ascension < 4 {
		return nil
	}
	done := false
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if done {
			return
		}
		done = true

		c.a4Counter++
		if c.a4Counter == 4 {
			c.a4Counter = 0
			if c.Cooldown(action.ActionSkill) > 0 {
				c.ReduceActionCooldown(action.ActionSkill, 60)
			}
		}
	}
}
