package xianyun

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const a4ICDKey = "xianyun-a4-icd"
const a4WindowKey = "xianyun-a4-window"
const a1Key = "xianyun-a1"
const a1Dur = 20 * 60

var a1Crit = []float64{0.0, 0.04, 0.06, 0.08, 0.10}

// 固有天賦1: 鶴雲波が敵に命中するたびに、
// チーム全員がブーストを1スタック獲得（20秒間、最大4スタック）。
// ブーストは落下攻撃ダメージの会心率を4%/6%/8%/10%上昇させ、
// 各スタックの持続時間は独立して計算される。

func (c *char) a1() {
	for i, char := range c.Core.Player.Chars() {
		mCR := make([]float64, attributes.EndStatType)
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("xianyun-a1-buff", -1),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				if atk.Info.AttackTag != attacks.AttackTagPlunge {
					return nil, false
				}
				stackCount := min(c.a1Buffer[i], 4)
				if stackCount == 0 {
					return nil, false
				}
				mCR[attributes.CR] = a1Crit[stackCount]
				return mCR, true
			},
		})
	}
}

func (c *char) a1cb() combat.AttackCBFunc {
	if c.Base.Ascension < 1 {
		return nil
	}
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}

		for i, char := range c.Core.Player.Chars() {
			idx := i
			c.a1Buffer[idx] += 1
			char.AddStatus(a1Key, a1Dur, true)
			char.SetTag(a1Key, min(c.a1Buffer[idx], 4))
			char.QueueCharTask(func() {
				// タグは現在結果UIには表示されない
				// ユーザーは .char.tags.xianyun-a1 でアクセス可能
				c.a1Buffer[idx] -= 1
				char.SetTag(a1Key, min(c.a1Buffer[idx], 4))
			}, a1Dur)
		}
	}
}

func (c *char) a4StartUpdate() {
	if c.Base.Ascension < 4 {
		return
	}

	c.a4src = c.Core.F
	c.a4AtkUpdate(c.Core.F)()
}

func (c *char) a4AtkUpdate(src int) func() {
	return func() {
		if c.a4src != src {
			return
		}
		c.a4Atk = c.TotalAtk()
		c.Core.Tasks.Add(c.a4AtkUpdate(src), 0.5*60)
	}
}

// 固有天賦4: 星と月の夕べが生成した星翼が仙助スタックを持つとき、
// 近くのアクティブキャラクターの落下攻撃衝撃波ダメージが閑雲の攻撃力の200%分増加する。
// この方法で得られる最大ダメージ増加量は9000。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}

	c.a4Max = 9000
	c.a4Ratio = 2.0

	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)
		if ae.Info.AttackTag != attacks.AttackTagPlunge {
			return false
		}

		// 衝突は元素量0。衝突ダメージにはバフを適用しない
		if ae.Info.Durability == 0 {
			return false
		}

		if !c.StatusIsActive(a4WindowKey) {
			return false
		}

		if c.StatusIsActive(a4ICDKey) {
			return false
		}

		// 固有天賦4の上限
		amt := min(c.a4Max, c.a4Ratio*c.a4Atk)

		c.Core.Log.NewEvent("Xianyun A4 proc dmg add", glog.LogPreDamageMod, ae.Info.ActorIndex).
			Write("atk", c.a4Atk).
			Write("ratio", c.a4Ratio).
			Write("addition", amt)

		ae.Info.FlatDmg += amt
		c.AddStatus(a4ICDKey, 0.4*60, true)

		return false
	}, "xianyun-starwicker-hook")
}
