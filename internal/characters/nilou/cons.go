package nilou

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 七域の舞が以下のように強化される:
// ・光幻のダメージ65%アップ
// ・水月のオーラの持続時間が6秒延長
func (c *char) c1() {
	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = 0.65

	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("nilou-c1", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.Abil != "Luminous Illusion" {
				return nil, false
			}
			return m, true
		},
	})
}

// 金杯の恭福の影響下のキャラが敵に水元素ダメージを与えた後、その敵の水元素耐性が10秒間35%低下する。
// 開花反応のダメージが敵に当たった後、その敵の草元素耐性が10秒間35%低下する。
// 固有天賦「庭園の舞踊」の解放が必要。
func (c *char) c2() {
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		dmg := args[2].(float64)
		t, ok := args[0].(*enemy.Enemy)
		if !ok {
			return false
		}
		if dmg == 0 {
			return false
		}

		char := c.Core.Player.ByIndex(atk.Info.ActorIndex)
		if !char.StatusIsActive(a1Status) {
			return false
		}

		if atk.Info.Element == attributes.Hydro {
			t.AddResistMod(combat.ResistMod{
				Base:  modifier.NewBaseWithHitlag("nilou-c2-hydro", 10*60),
				Ele:   attributes.Hydro,
				Value: -0.35,
			})
		} else if atk.Info.AttackTag == attacks.AttackTagBloom {
			t.AddResistMod(combat.ResistMod{
				Base:  modifier.NewBaseWithHitlag("nilou-c2-dendro", 10*60),
				Ele:   attributes.Dendro,
				Value: -0.35,
			})
		}

		return false
	}, "nilou-c2")
}

// 七域の舞のピルエット3歩目が敵に命中した後、ニィロウは元素エネルギーを15獲得し、
// 元素爆発のダメージが8秒間50%アップする。
func (c *char) c4() {
	c.AddEnergy("nilou-c4", 15)

	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = 0.5
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBaseWithHitlag("nilou-c4", 8*60),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagElementalBurst {
				return nil, false
			}
			return m, true
		},
	})
}

func (c *char) c4cb() combat.AttackCBFunc {
	if c.Base.Cons < 4 {
		return nil
	}
	if c.Tag(skillStep) != 2 || !c.StatusIsActive(pirouetteStatus) {
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
		c.c4()
		done = true
	}
}

// 最大HP1,000ポイントごとに、ニィロウの会心率0.6%と会心ダメージ1.2%がアップする。
// 最大会心率30%、最大会心ダメージ60%。
func (c *char) c6() {
	// NoStat属性によるスタックオーバーフローを避けるため、CRとCDを分離
	mCR := make([]float64, attributes.EndStatType)
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("nilou-c6-cr", -1),
		AffectedStat: attributes.CR,
		Extra:        true,
		Amount: func() ([]float64, bool) {
			cr := c.MaxHP() * 0.001 * 0.006
			if cr > 0.3 {
				cr = 0.3
			}
			mCR[attributes.CR] = cr
			return mCR, true
		},
	})

	mCD := make([]float64, attributes.EndStatType)
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("nilou-c6-cd", -1),
		AffectedStat: attributes.CD,
		Extra:        true,
		Amount: func() ([]float64, bool) {
			cd := c.MaxHP() * 0.001 * 0.012
			if cd > 0.6 {
				cd = 0.6
			}
			mCD[attributes.CD] = cd
			return mCD, true
		},
	})

	c.QueueCharTask(c.c6, 60)
}
