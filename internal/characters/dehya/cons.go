package dehya

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// ディヘヤのHP上限が20%増加する。また、以下の攻撃を行う際、HP上限に基づくダメージボーナスを獲得する：
// ・熔鉄の炎のダメージがHP上限の3.6%分増加する。
// ・獅子拳のダメージがHP上限の6%分増加する。
func (c *char) c1() {
	if c.Base.Cons < 1 {
		return
	}
	// HP20%増加
	m := make([]float64, attributes.EndStatType)
	m[attributes.HPP] = 0.2
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("dehya-c1", -1),
		AffectedStat: attributes.HPP,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})
	// アビリティのフラットダメージ
	c.c1FlatDmgRatioE = 0.036
	c.c1FlatDmgRatioQ = 0.06
}

// ディヘヤが熔鉄の炎・烈炎潜伝を使用した際、再生成される浄焔の領域の持続時間が6秒延長される。
func (c *char) c2IncreaseDur() {
	if c.Base.Cons < 2 {
		return
	}
	c.sanctumSavedDur += 360
}

// また、浄焔の領域がフィールドに存在する時、領域内のアクティブキャラクターが攻撃を受けると、
// 次の追撃ダメージが50%増加する。
func (c *char) c2() {
	if c.Base.Cons < 2 {
		return
	}
	val := make([]float64, attributes.EndStatType)
	val[attributes.DmgP] = 0.5
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("dehya-sanctum-dot-c2", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.Abil != skillDoTAbil || !c.hasC2DamageBuff {
				return nil, false
			}
			return val, true
		},
	})
	c.Core.Events.Subscribe(event.OnPlayerHit, func(args ...interface{}) bool {
		char := args[0].(int)
		// アクティブキャラがヒットされていなければ発動しない
		if char != c.Core.Player.Active() {
			return false
		}
		// 領域がアクティブである必要がある
		if !c.StatusIsActive(dehyaFieldKey) {
			return false
		}
		// プレイヤーが領域内にいる必要がある
		if !c.Core.Combat.Player().IsWithinArea(c.skillArea) {
			return false
		}
		c.Core.Log.NewEvent("dehya-sanctum-c2-damage activated", glog.LogCharacterEvent, c.Index)
		c.hasC2DamageBuff = true
		return false
	}, "dehya-c2")
}

// 獅子拳中に発動する炎鬣の拳と焼尽の駆動が敵に命中した際、
// ディヘヤのエネルギーを1.5回復し、HP上限の2.5%分のHPを回復する。この効果は0.2秒毎に1回発動可能。
const c4Key = "dehya-c4"
const c4ICDKey = "dehya-c4-icd"

func (c *char) c4CB() combat.AttackCBFunc {
	if c.Base.Cons < 4 {
		return nil
	}
	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if c.StatusIsActive(c4ICDKey) {
			return
		}
		c.AddStatus(c4ICDKey, 0.2*60, true)

		c.AddEnergy(c4Key, 1.5)
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index,
			Target:  c.Index,
			Message: "An Oath Abiding (C4)",
			Src:     0.025 * c.MaxHP(),
			Bonus:   c.Stat(attributes.Heal),
		})
	}
}

// 獅子拳の会心率が10%上昇する。
// また、一回の炎鬣獅子状態中に炎鬣の拳が敵に命中し会心が発生すると、
// 獅子拳の会心ダメージが残りの炎鬣獅子の持続時間中15%増加し、持続時間が0.5秒延長される。
// この効果は0.2秒毎に1回発動可能。持続時間は最大2秒延長可能、会心ダメージは最大60%まで増加可能。
func (c *char) c6() {
	if c.Base.Cons < 6 {
		return
	}
	val := make([]float64, attributes.EndStatType)
	val[attributes.CR] = 0.1

	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("dehya-c6", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagElementalBurst {
				return nil, false
			}
			val[attributes.CD] = 0.15 * float64(c.c6Count)

			return val, true
		},
	})
}

const c6ICDKey = "dehya-c6-icd"

func (c *char) c6CB() combat.AttackCBFunc {
	if c.Base.Cons < 6 {
		return nil
	}

	return func(a combat.AttackCB) {
		if a.Target.Type() != targets.TargettableEnemy {
			return
		}
		if !a.IsCrit {
			return
		}
		if c.c6Count == 4 {
			return
		}
		if c.StatusIsActive(c6ICDKey) {
			return
		}
		c.AddStatus(c6ICDKey, 0.2*60, true)

		c.c6Count++
		c.ExtendStatus(burstKey, 0.5*60)
	}
}
