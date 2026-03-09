package flins

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	c1EnergyICD    = "flins-c1-energy-icd"
	c2NorthlandKey = "flins-c2-northland"
	c2ResDebuff    = "flins-c2-res-debuff"
)

// 1命ノ星座
// 特殊元素スキル「北地の槍嵐」の基本クールダウンが4秒に短縮される。
// また、パーティメンバーがルナチャージ反応を発動すると、Flinsの元素エネルギーが8回復する。この効果は5.5秒に1回発動可能。
func (c *char) c1() {
	if c.Base.Cons < 1 {
		c.northlandCD = 6 * 60
		return
	}
	c.northlandCD = 4 * 60

	// ルナチャージ反応イベントを購読してエネルギー回復
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.AttackTag != attacks.AttackTagLCDamage {
			return false
		}
		if c.StatusIsActive(c1EnergyICD) {
			return false
		}
		c.AddEnergy("flins-c1", 8)
		c.AddStatus(c1EnergyICD, 5.5*60, true)
		return false
	}, "flins-c1-energy")
}

// 2命ノ星座
// 特殊元素スキル「北地の槍嵐」使用後6秒間、Flinsの次の通常攻撃が敵に命中すると、Flinsの攻撃力の50%に相当する雷元素範囲ダメージを追加で与える。このダメージはルナチャージダメージとみなされる。
// ムーンサインが「ムーンサイン: 昇詼の輝き」の場合、Flinsがフィールド上にいる間、雷元素攻撃が敵に命中すると、その敵の雷元素耐性が7秒間25%減少する。
func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}

	// パート1: 北地の槍嵐後の追加ダメージ
	// c2NorthlandKeyステータスをチェックするコールバックで処理
	// コールバックはattackE()関数内で通常攻撃に追加される

	// パート2: 「ムーンサイン: 昇詼の輝き」時の雷元素耐性減少
	if !c.MoonsignAscendant {
		return
	}

	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != c.Index {
			return false
		}
		if atk.Info.Element != attributes.Electro {
			return false
		}
		if c.Core.Player.Active() != c.Index {
			return false
		}

		enemy, ok := args[0].(combat.Enemy)
		if !ok {
			return false
		}
		enemy.AddResistMod(combat.ResistMod{
			Base:  modifier.NewBaseWithHitlag(c2ResDebuff, 7*60),
			Ele:   attributes.Electro,
			Value: -0.25,
		})
		return false
	}, "flins-c2-res")
}

// 2命ノ星座: 北地の槍嵐後の追加ダメージコールバック
func (c *char) c2AdditionalDamage() combat.AttackCBFunc {
	if c.Base.Cons < 2 {
		return nil
	}

	done := false
	return func(a combat.AttackCB) {
		if done {
			return
		}
		if !c.StatusIsActive(c2NorthlandKey) {
			return
		}
		_, ok := a.Target.(combat.Enemy)
		if !ok {
			return
		}

		done = true
		c.DeleteStatus(c2NorthlandKey)

		// 攻撃力の50%を雷元素ダメージとして追加（ルナチャージ）
		ai := combat.AttackInfo{
			ActorIndex: c.Index,
			Abil:       "Flins C2 Dummy",
			FlatDmg:    0,
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 99), 0, 0)
	}
}

// 4命ノ星座
// Flinsの攻撃力が20%増加する。
// また、固有天賦「囁きの炎」が変更される: Flinsの元素熟知が攻撃力の10%分増加する。この方法で得られる最大増加量は220。
func (c *char) c4() {
	if c.Base.Cons < 4 {
		return
	}

	// 攻撃力20%増加
	m := make([]float64, attributes.EndStatType)
	c.AddStatMod(character.StatMod{
		Base: modifier.NewBaseWithHitlag("Flins C4 ATK", -1),
		Amount: func() ([]float64, bool) {
			m[attributes.ATKP] = 0.20
			return m, true
		},
	})
}

// 6命ノ星座
// Flinsのルナチャージ反応が敵に与えるダメージが35%増加する。
// ムーンサインが「ムーンサイン: 昇詼の輝き」の場合、付近のパーティメンバー全員のルナチャージダメージが10%増加する。
func (c *char) c6() {
	if c.Base.Cons < 6 {
		return
	}

	// Flins自身のルナチャージダメージボーナス: 35%
	c.AddElevationMod(character.ElevationMod{
		Base: modifier.NewBase("Flins C6", -1),
		Amount: func(ai combat.AttackInfo) (float64, bool) {
			if ai.AttackTag == attacks.AttackTagLCDamage {
				return 0.35, false
			} else {
				return 0, false
			}
		},
	})

	// 「ムーンサイン: 昇詼の輝き」時のチーム全体ルナチャージダメージボーナス: 10%
	if c.MoonsignAscendant {
		for _, char := range c.Core.Player.Chars() {
			char.AddElevationMod(character.ElevationMod{
				Base: modifier.NewBase("Flins C6 Team", -1),
				Amount: func(ai combat.AttackInfo) (float64, bool) {
					if ai.AttackTag == attacks.AttackTagLCDamage {
						return 0.1, false
					} else {
						return 0, false
					}
				},
			})
		}
	}
}
