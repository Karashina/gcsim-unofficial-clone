package nilou

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/reactable"
)

const (
	a1Status = "nilou-a1"
	a4Mod    = "nilou-a4"
)

// パーティー全員が草元素または水元素で、かつ草元素キャラと水元素キャラがそれぞれ1人以上いる場合:
// ニィロウの七域の舞の3回目のステップ完了時、近くの全キャラクターに
// 金杯の恭福を30秒間付与する。
// 金杯の恭福の影響下のキャラクターが草元素攻撃を受けた時、近くの全キャラの元素熔合が10秒間100アップする。
// また、開花反応時に草元素コアではなく豊穣のコアを生成する。
// 豊穣のコアは生成後すぐに爆発し、より広い範囲を持つ。
// 豊穣のコアは超開花や烈開花をトリガーできず、草元素コアと上限を共有する。豊穣のコアのダメージは
// 開花反応による草元素コアのダメージとみなされる。
// パーティーがこの固有天賦の条件を満たさない場合、既存の金杯の恭福効果は取り消される。
func (c *char) a1() {
	if c.Base.Ascension < 1 || !c.onlyBloomTeam {
		return
	}

	for _, this := range c.Core.Player.Chars() {
		this.AddStatus(a1Status, 30*60, true)
	}
	c.a4()

	// 豊穣のコア
	c.Core.Events.Subscribe(event.OnDendroCore, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		char := c.Core.Player.ByIndex(atk.Info.ActorIndex)
		if !char.StatusIsActive(a1Status) {
			return false
		}
		g, ok := args[0].(*reactable.DendroCore)
		if !ok {
			return false
		}

		b := newBountifulCore(c.Core, g.Gadget.Pos(), atk)
		b.Gadget.SetKey(g.Gadget.Key())
		c.Core.Combat.ReplaceGadget(g.Key(), b)
		// 爆発を防止
		g.Gadget.OnExpiry = nil
		g.Gadget.OnKill = nil

		return false
	}, "nilou-a1-cores")

	c.Core.Events.Subscribe(event.OnPlayerHit, func(args ...interface{}) bool {
		charIndex := args[0].(int)
		char := c.Core.Player.ByIndex(charIndex)
		if !char.StatusIsActive(a1Status) {
			return false
		}
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.Element != attributes.Dendro {
			return false
		}

		m := make([]float64, attributes.EndStatType)
		m[attributes.EM] = 100
		for _, this := range c.Core.Player.Chars() {
			this.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag("nilou-a1-em", 10*60),
				AffectedStat: attributes.EM,
				Amount: func() ([]float64, bool) {
					return m, true
				},
			})
		}

		return false
	}, "nilou-a1")
}

// ニィロウの最大HPが30,000を超えた分、1,000ポイントごとに
// 金杯の恭福の影響下のキャラが生成した豊穣のコアのダメージが9%増加する。
// 最大増加は400%。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	for _, this := range c.Core.Player.Chars() {
		// TODO: a4は追加バフであるべき
		this.AddReactBonusMod(character.ReactBonusMod{
			Base: modifier.NewBaseWithHitlag(a4Mod, 30*60),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				if ai.AttackTag != attacks.AttackTagBloom {
					return 0, false
				}
				if ai.ICDTag != attacks.ICDTagBountifulCoreDamage {
					return 0, false
				}

				c.Core.Combat.Log.NewEvent("adding nilou a4 bonus", glog.LogCharacterEvent, c.Index).Write("bonus", c.a4Bonus)
				return c.a4Bonus, false
			},
		})
	}
	c.a4Src = c.Core.F
	c.QueueCharTask(c.updateA4Bonus(c.a4Src), 0.5*60)
}

func (c *char) updateA4Bonus(src int) func() {
	return func() {
		if c.a4Src != src {
			return
		}
		if !c.ReactBonusModIsActive(a4Mod) {
			return
		}

		c.a4Bonus = (c.MaxHP() - 30000) * 0.001 * 0.09
		if c.a4Bonus < 0 {
			c.a4Bonus = 0
		} else if c.a4Bonus > 4 {
			c.a4Bonus = 4
		}

		c.QueueCharTask(c.updateA4Bonus(src), 0.5*60)
	}
}
