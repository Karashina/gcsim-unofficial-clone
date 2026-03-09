package yanfei

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 煙绯が重撃で朱印を消費すると、
// 朱印1枚ごとに煙绯の炎元素ダメージバフが5%増加する。
// この効果は6秒間持続する。効果持続中に再度重撃を使用すると、
// 前の効果が解除される。
func (c *char) a1(stacks int) {
	if c.Base.Ascension < 1 {
		return
	}
	c.a1Buff[attributes.PyroP] = float64(stacks) * 0.05
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag("yanfei-a1", 360),
		AffectedStat: attributes.PyroP,
		Amount: func() ([]float64, bool) {
			return c.a1Buff, true
		},
	})
}

// 煙绯の重撃が敵に会心ヒットした場合、
// 攻撃力80%の追加の範囲炎元素ダメージを与える。
// このダメージは重撃ダメージとみなされる。
func (c *char) makeA4CB() combat.AttackCBFunc {
	if c.Base.Ascension < 4 {
		return nil
	}
	done := false
	return func(a combat.AttackCB) {
		trg := a.Target
		if trg.Type() != targets.TargettableEnemy {
			return
		}
		if c.Core.Player.Active() != c.Index {
			return
		}
		if !a.IsCrit {
			return
		}
		if done {
			return
		}
		done = true

		ai := combat.AttackInfo{
			ActorIndex:         c.Index,
			Abil:               "Blazing Eye (A4)",
			AttackTag:          attacks.AttackTagExtra,
			ICDTag:             attacks.ICDTagNone,
			ICDGroup:           attacks.ICDGroupDefault,
			StrikeType:         attacks.StrikeTypeDefault,
			Element:            attributes.Pyro,
			Durability:         25,
			Mult:               0.8,
			HitlagFactor:       0.05,
			CanBeDefenseHalted: true,
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(trg, nil, 3.5), 10, 10)
	}
}
