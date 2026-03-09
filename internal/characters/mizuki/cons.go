package mizuki

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

const (
	c1Key               = "mizuki-c1"
	c1Interval          = 3.5 * 60
	c1Duration          = 3 * 60
	c1Multiplier        = 11.0
	c1Range             = 12
	c2Key               = "mizuki-c2"
	c2EMMultiplier      = 0.0004
	c2Interval          = 0.5 * 60
	c4EnergyGenerations = 4
	c4Key               = "mizuki-c4"
	c4Energy            = 5
	c6Key               = "mizuki-c6"
	c6CR                = 0.3
	c6CD                = 1.0
)

// 夢見月瑞希がDreamdrifter状態中、3.5秒ごとに周囲の敵に3秒間「二十三夜の待望」効果を継続的に付与する。
// 前述の効果がアクティブな間に風元素ダメージによる拡散反応を受けた敵は、
// その効果がキャンセルされ、この拡散のその敵に対するダメージが
// 瑞希の元素熟知の1100%分増加する。
func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}

	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		e, ok := args[0].(*enemy.Enemy)
		atk := args[1].(*combat.AttackEvent)
		if !ok {
			return false
		}

		// 敵がデバフを持っているか確認
		if !e.StatusIsActive(c1Key) {
			return false
		}

		// 拡散のみ。拡散の発生源は問わず、瑞希でも他の風元素キャラでもよい。
		switch atk.Info.AttackTag {
		case attacks.AttackTagSwirlCryo:
		case attacks.AttackTagSwirlElectro:
		case attacks.AttackTagSwirlHydro:
		case attacks.AttackTagSwirlPyro:
		default:
			return false
		}

		// 0ダメージの拡散では発動しない（例：水元素範囲拡散や拡散ICD）
		if atk.Info.FlatDmg == 0 {
			return false
		}

		additionalDmg := c1Multiplier * c.c1EM

		c.Core.Log.NewEvent("mizuki c1 proc", glog.LogPreDamageMod, atk.Info.ActorIndex).
			Write("before", atk.Info.FlatDmg).
			Write("addition", additionalDmg).
			Write("final", atk.Info.FlatDmg+additionalDmg)

		atk.Info.FlatDmg += additionalDmg
		atk.Info.Abil += " (Mizuki C1)"

		// 効果をキャンセル
		e.DeleteStatus(c1Key)

		return false
	}, c1Key)
}

func (c *char) c1Task(src, hitmark int) {
	c.QueueCharTask(func() {
		if c.cloudSrc != src {
			return
		}
		if !c.StatusIsActive(dreamDrifterStateKey) {
			return
		}

		c.c1EM = c.Stat(attributes.EM)
		area := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, c1Range)
		for _, target := range c.Core.Combat.EnemiesWithinArea(area, nil) {
			if e, ok := target.(*enemy.Enemy); ok {
				// ヒットラグの影響を受けているか確認する方法はあるのか？
				e.AddStatus(c1Key, c1Duration, true)
			}
		}
		c.c1Task(src, c1Interval)
	}, hitmark)
}

// 夢見月瑞希がDreamdrifter状態に入った時、彼女の元素熟知1ポイントにつき、
// 周囲のパーティメンバー全員の炎・水・氷・雷元素ダメージボーナスが0.04%増加する（Dreamdrifter状態終了まで）。
func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}

	c.c2Buff = make([]float64, attributes.EndStatType)
	c.c2UpdateTask()

	for _, char := range c.Core.Player.Chars() {
		if char.Index == c.Index {
			continue
		}
		// TODO: 2凸を入手したら、これが本当に静的バフかテストする
		char.AddStatMod(character.StatMod{
			Base: modifier.NewBase(c2Key, -1),
			Amount: func() ([]float64, bool) {
				if !c.StatusIsActive(dreamDrifterStateKey) {
					return nil, false
				}
				return c.c2Buff, true
			},
		})
	}
}

func (c *char) c2UpdateTask() {
	if c.Base.Cons < 2 {
		return
	}

	c.QueueCharTask(func() {
		dmgBonus := c.NonExtraStat(attributes.EM) * c2EMMultiplier
		c.c2Buff[attributes.PyroP] = dmgBonus
		c.c2Buff[attributes.HydroP] = dmgBonus
		c.c2Buff[attributes.ElectroP] = dmgBonus
		c.c2Buff[attributes.CryoP] = dmgBonus

		c.c2UpdateTask()
	}, c2Interval)
}

// 元素爆発「安楽秘湯浴」の夢見式特製おやつを拾うとダメージと回復の両方が発生し、
// 夢見月瑞希に元素エネルギーを5回復する。この方法でのエネルギー回復は
// 安楽秘湯浴1回につき最大4回まで。
func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}

	if c.c4EnergyGenerationsRemaining > 0 {
		c.c4EnergyGenerationsRemaining--
		c.AddEnergy(c4Key, c4Energy)
	}
}

// 夢見月瑞希がDreamdrifter状態中、周囲のパーティメンバーの拡散ダメージが会心可能になる。
// 会心率は30%固定、会心ダメージは100%固定。
func (c *char) c6() {
	if c.Base.Cons < 6 {
		return
	}

	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		_, ok := args[0].(*enemy.Enemy)
		if !ok {
			return false
		}

		ae := args[1].(*combat.AttackEvent)

		// 拡散のみ。拡散の発生源は問わず、瑞希でも他の風元素キャラでもよい。
		switch ae.Info.AttackTag {
		case attacks.AttackTagSwirlPyro:
		case attacks.AttackTagSwirlCryo:
		case attacks.AttackTagSwirlHydro:
		case attacks.AttackTagSwirlElectro:
		default:
			return false
		}

		// 瑞希がDreamdrifter状態中のみ効果あり
		if !c.StatusIsActive(dreamDrifterStateKey) {
			return false
		}

		// 会心率/ダメージは30% CR、100% CDに固定
		ae.Snapshot.Stats[attributes.CR] = c6CR
		ae.Snapshot.Stats[attributes.CD] = c6CD

		return false
	}, c6Key)
}
