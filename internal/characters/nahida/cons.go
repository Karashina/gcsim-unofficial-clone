package nahida

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// マヤの祠殿を解き放つときにチームメンバーの元素タイプを集計する際、
// 炎・雷・水のキャラクター数がそれぞれ1加算される。
func (c *char) c1() {
	c.pyroCount++
	c.hydroCount++
	c.electroCount++
}

// ナヒーダ自身のスカンダの種でマークされた敵に以下の効果が適用される:
//   - 燃焼・開花・超開花・烈開花の反応ダメージが会心できる。
//     会心率と会心ダメージはそれぞれ20%と100%に固定。
//   - 激化・激化促進・拡散の影響を受けて8秒以内に、防御力が30%低下する。
func (c *char) c2() {
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		t, ok := args[0].(*enemy.Enemy)
		if !ok {
			return false
		}
		ae := args[1].(*combat.AttackEvent)

		switch ae.Info.AttackTag {
		case attacks.AttackTagBurningDamage:
		case attacks.AttackTagBloom:
		case attacks.AttackTagHyperbloom:
		case attacks.AttackTagBurgeon:
		default:
			return false
		}

		if !t.StatusIsActive(skillMarkKey) {
			return false
		}

		//TODO: 本当に += でよいのか？
		ae.Snapshot.Stats[attributes.CR] += 0.2
		ae.Snapshot.Stats[attributes.CD] += 1

		c.Core.Log.NewEvent("nahida c2 buff", glog.LogCharacterEvent, ae.Info.ActorIndex).
			Write("final_crit", ae.Snapshot.Stats[attributes.CR])

		return false
	}, "nahida-c2-reaction-dmg-buff")

	cb := func(rx event.Event) event.Hook {
		return func(args ...interface{}) bool {
			t, ok := args[0].(*enemy.Enemy)
			if !ok {
				return false
			}
			if !t.StatusIsActive(skillMarkKey) {
				return false
			}
			t.AddDefMod(combat.DefMod{
				Base:  modifier.NewBaseWithHitlag("nahida-c2", 480),
				Value: -0.3,
			})
			return false
		}
	}

	c.Core.Events.Subscribe(event.OnQuicken, cb(event.OnQuicken), "nahida-c2-def-shred-quicken")
	c.Core.Events.Subscribe(event.OnAggravate, cb(event.OnAggravate), "nahida-c2-def-shred-aggravate")
	c.Core.Events.Subscribe(event.OnSpread, cb(event.OnSpread), "nahida-c2-def-shred-spread")
}

// 近くの敵が蓮生の洞察のスカンダの種に影響を受けている数が
// 1/2/3/(4以上)のとき、ナヒーダの元素熔研が
// 100/120/140/160上昇する。
func (c *char) c4() {
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("nahida-c4", -1),
		AffectedStat: attributes.EM,
		Amount: func() ([]float64, bool) {
			enemies := c.Core.Combat.EnemiesWithinArea(
				combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 30),
				func(t combat.Enemy) bool {
					return t.StatusIsActive(skillMarkKey)
				},
			)
			count := len(enemies)
			if count > 4 {
				count = 4
			}
			if count == 0 {
				return nil, false
			}
			c.c4Buff[attributes.EM] = float64(80 + count*20)
			return c.c4Buff, true
		},
	})
}

const (
	c6ICDKey    = "nahida-c6-icd"
	c6ActiveKey = "nahida-c6"
)

// 幻心を解き放った後、蓮生の洞察のスカンダの種の影響を受けている敵に
// ナヒーダが通常攻撃または重撃を命中させると、その敵と接続された全ての敵に
// 三浄浄化・業障基滅を発動し、ナヒーダの攻撃力の200%と元素熔研の400%に基づく
// 草元素ダメージを与える。三浄浄化・業障基滅のダメージは元素スキルダメージ扱いで、
// 0.2秒に1回発動可能。この効果は最大10秒間持続し、ナヒーダが6回
// 三浄浄化・業障基滅を発動した後に解除される。
func (c *char) makeC6CB() combat.AttackCBFunc {
	if c.Base.Cons < 6 {
		return nil
	}
	return func(a combat.AttackCB) {
		e, ok := a.Target.(*enemy.Enemy)
		if !ok {
			return
		}
		if !e.StatusIsActive(skillMarkKey) {
			return
		}
		if c.c6Count >= 6 {
			return
		}
		if !c.StatusIsActive(c6ActiveKey) {
			return
		}
		if c.StatusIsActive(c6ICDKey) {
			return
		}
		c.AddStatus(c6ICDKey, 0.2*60, true)

		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Tri-Karma Purification: Karmic Oblivion",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagNahidaC6,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Dendro,
			Durability: 25,
			Mult:       2,
		}
		snap := c.Snapshot(&ai)
		ai.FlatDmg = snap.Stats[attributes.EM] * 4
		for _, v := range c.Core.Combat.Enemies() {
			e, ok := v.(*enemy.Enemy)
			if !ok {
				continue
			}
			if !e.StatusIsActive(skillMarkKey) {
				continue
			}
			c.Core.QueueAttackWithSnap(
				ai,
				snap,
				combat.NewSingleTargetHit(e.Key()),
				1,
			)
		}

		c.c6Count++
	}
}
