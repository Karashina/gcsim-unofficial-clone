package mona

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const c6Key = "mona-c6"

// 1命ノ星座:
// 自身のパーティメンバーが星異の影響を受けた敵を攻撃すると、水元素関連の元素反応の効果が8秒間強化される:
// - 感電ダメージが15%増加。
// - 蒸発ダメージが15%増加。
// - 水元素拡散ダメージが15%増加。
// - 凍結時間が15%延長。
func (c *char) c1() {
	// TODO: 「凍結時間が15%延長」はバグがある
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		// ターゲットにデバフがなければ無視
		t, ok := args[0].(*enemy.Enemy)
		if !ok {
			return false
		}
		if !t.StatusIsActive(bubbleKey) && !t.StatusIsActive(omenKey) {
			return false
		}
		// 全パーティメンバーにc1を追加、1フレーム遅延。理由:
		// 「このボーナスはトリガーとなった攻撃には適用されず、星天帰還の虚実の泡による水元素ダメージにも、元素反応による場合を含めて適用されない。」
		for _, x := range c.Core.Player.Chars() {
			char := x
			c.Core.Tasks.Add(func() {
				// TODO: 「蒸発ダメージが15%増加」はスナップショットされるべき。参照: https://library.keqingmains.com/evidence/characters/hydro/mona#mona-c1-snapshot-for-vape
				// ReactBonusMod のリファクタリングが必要
				char.AddReactBonusMod(character.ReactBonusMod{
					Base: modifier.NewBase("mona-c1", 8*60),
					Amount: func(ai combat.AttackInfo) (float64, bool) {
						// フィールド外では機能しない
						if c.Core.Player.Active() != char.Index {
							return 0, false
						}
						// 感電ダメージが15%増加。
						if ai.AttackTag == attacks.AttackTagECDamage {
							return 0.15, false
						}
						// 蒸発ダメージが15%増加。
						// 水元素拡散が蒸発を起こすのはAoE水元素拡散経由のみで、そもそもダメージは発生しないため問題ない
						if ai.Amped {
							return 0.15, false
						}
						// 水元素拡散ダメージが15%増加。
						if ai.AttackTag == attacks.AttackTagSwirlHydro {
							return 0.15, false
						}
						return 0, false
					},
				})
				char.AddLCReactBonusMod(character.LCReactBonusMod{
					Base: modifier.NewBase("mona-c1-lc", 8*60),
					Amount: func(ai combat.AttackInfo) (float64, bool) {
						return 0.15, false
					},
				})
			}, 1)
		}
		return false
	}, "mona-c1-check")
}

// 2凸:
// 通常攻撃命中時、20%の確率で自動的に重撃が続く。
// この効果は5秒ごとに1回のみ発動。
func (c *char) c2(a combat.AttackCB) {
	trg := a.Target
	if c.Base.Cons < 2 {
		return
	}
	if a.Target.Type() != targets.TargettableEnemy {
		return
	}
	if c.Core.Rand.Float64() > .2 {
		return
	}
	if c.c2icd > c.Core.F {
		return
	}
	c.c2icd = c.Core.F + 300 // 5秒ごと
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Charge Attack",
		AttackTag:  attacks.AttackTagExtra,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
		Mult:       charge[c.TalentLvlAttack()],
	}

	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(trg, nil, 3), 0, 0)
}

// 4凸:
// パーティメンバーが星異の影響を受けた敵を攻撃すると、会心率が15%増加する。
func (c *char) c4() {
	m := make([]float64, attributes.EndStatType)
	m[attributes.CR] = 0.15

	for _, char := range c.Core.Player.Chars() {
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("mona-c4", -1),
			Amount: func(_ *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				x, ok := t.(*enemy.Enemy)
				if !ok {
					return nil, false
				}
				// 泡または星異のどちらかが存在する場合のみ有効
				if x.StatusIsActive(bubbleKey) || x.StatusIsActive(omenKey) {
					return m, true
				}
				return nil, false
			},
		})
	}
}

// 6命ノ星座:
// 虚実流動に入ると、モナは移動1秒ごとに次の重撃ダメージが60%増加する。
// この方法で得られるダメージバフは最大180%。
// 効果は最大8秒間持続する。
func (c *char) c6(src int) func() {
	return func() {
		if c.c6Src != src {
			c.Core.Log.NewEvent(fmt.Sprintf("%v stack gain check ignored, src diff", c6Key), glog.LogCharacterEvent, c.Index).
				Write("src", src).
				Write("new src", c.c6Src)
			return
		}
		// モナでなければ何もしない
		if c.Core.Player.Active() != c.Index {
			return
		}
		// ダッシュ中でなければ何もしない
		if c.Core.Player.CurrentState() != action.DashState {
			return
		}

		c.c6Stacks++
		if c.c6Stacks > 3 {
			c.c6Stacks = 3
		}
		c.Core.Log.NewEvent(fmt.Sprintf("%v stack gained", c6Key), glog.LogCharacterEvent, c.Index).
			Write("c6Stacks", c.c6Stacks)

		m := make([]float64, attributes.EndStatType)
		c.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase(c6Key, 8*60),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				if atk.Info.AttackTag != attacks.AttackTagExtra {
					return nil, false
				}
				m[attributes.DmgP] = 0.60 * float64(c.c6Stacks)
				return m, true
			},
		})

		// 重撃を使用しなかった場合、8秒で6凸スタックをリセット
		c.Core.Tasks.Add(c.c6TimerReset, 8*60+1)
		// 1秒後に次のスタックとバフ更新をキューに追加
		c.Core.Tasks.Add(c.c6(src), 60)
	}
}

func (c *char) makeC6CAResetCB() combat.AttackCBFunc {
	if c.Base.Cons < 6 || !c.StatusIsActive(c6Key) {
		return nil
	}
	return func(a combat.AttackCB) {
		if a.Target.Type() == targets.TargettableEnemy {
			return
		}
		if !c.StatusIsActive(c6Key) {
			return
		}
		c.DeleteStatus(c6Key)
		c.c6Stacks = 0
		c.Core.Log.NewEvent(fmt.Sprintf("%v stacks reset via charge attack", c6Key), glog.LogCharacterEvent, c.Index)
	}
}

func (c *char) c6TimerReset() {
	// c6バフが切れる前に重撃を使わなかった場合にC6スタックをリセット
	if c.c6Stacks > 0 && !c.StatusIsActive(c6Key) {
		c.c6Stacks = 0
		c.Core.Log.NewEvent(fmt.Sprintf("%v stacks reset via timer", c6Key), glog.LogCharacterEvent, c.Index)
	}
}
