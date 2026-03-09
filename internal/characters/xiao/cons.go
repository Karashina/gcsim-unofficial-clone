package xiao

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 魉の2凸を実装:
// パーティにいるがフィールドにいない時、元素チャージ効率+25%
func (c *char) c2() {
	m := make([]float64, attributes.EndStatType)
	m[attributes.ER] = 0.25
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("xiao-c2", -1),
		AffectedStat: attributes.ER,
		Amount: func() ([]float64, bool) {
			if c.Core.Player.Active() != c.Index {
				return m, true
			}
			return nil, false
		},
	})
}

// 魉の4凸を実装:
// HPが50%以下の時、防御力+100%。
func (c *char) c4() {
	//TODO: ゲーム内では実際には0.3秒ごとのチェック。HPが50%未満ならバフ有効、
	// 次のチェックまで継続
	m := make([]float64, attributes.EndStatType)
	m[attributes.DEFP] = 1
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("xiao-c4", -1),
		AffectedStat: attributes.DEFP,
		Amount: func() ([]float64, bool) {
			if c.CurrentHPRatio() <= 0.5 {
				return m, true
			}
			return nil, false
		},
	})
}

const c6BuffKey = "xiao-c6"

// 魉の6凸を実装:
// 「妞降」の効果中、落下攻撃で2体以上の敵に命中すると、即座に風輪両立のチャージを1得る。
// その後1秒間、CDを無視して風輪両立を使用可能。
// OnDamage イベントチェッカーを追加 - 落下攻撃ダメージが2回以上記録されたらC6を発動
func (c *char) c6cb() combat.AttackCBFunc {
	if c.Base.Cons < 6 {
		return nil
	}
	c.c6Count = 0
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if !c.StatusIsActive(burstBuffKey) {
			return
		}
		if c.StatusIsActive(c6BuffKey) {
			return
		}
		c.c6Count++
		if c.c6Count == 2 {
			c.ResetActionCooldown(action.ActionSkill)
			// 連続3回のスキルをカバーするために1.2秒
			c.AddStatus(c6BuffKey, 72, true)
			c.Core.Log.NewEvent("Xiao C6 activated", glog.LogCharacterEvent, c.Index).
				Write("new E charges", c.Tags["eCharge"]).
				Write("expiry", c.Core.F+60)

			c.c6Count = 0
		}
	}
}
