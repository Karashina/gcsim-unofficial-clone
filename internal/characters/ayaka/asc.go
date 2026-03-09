package ayaka

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 神里流・氷華を使用後、神里綾華の通常攻撃と重撃のダメージが6秒間30%増加する。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = 0.3
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBaseWithHitlag("ayaka-a1", 360),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			return m, atk.Info.AttackTag == attacks.AttackTagNormal || atk.Info.AttackTag == attacks.AttackTagExtra
		},
	})
}

// 神里流・霰歩の終了時の氷元素付与が敵にヒットすると、神里綾華は以下の効果を得る：
//
// - スタミナ10回復
//
// - 10秒間、氷元素ダメージバフ18%を獲得。
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

		c.Core.Player.RestoreStam(10)

		m := make([]float64, attributes.EndStatType)
		m[attributes.CryoP] = 0.18
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("ayaka-a4", 600),
			AffectedStat: attributes.CryoP,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}
}
