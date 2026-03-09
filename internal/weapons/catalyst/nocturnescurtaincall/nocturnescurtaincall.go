package nocturnescurtaincall

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.NocturnesCurtainCall, NewWeapon)
}

type Weapon struct {
	Index int
	core  *core.Core
	char  *character.CharWrapper
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// Nocturne's Curtain Call（夜幕のカーテンコール）
// HP上限が10/12/14/16/18%増加する。
// Lunar反応を発動した時、または敵にLunar Reactionダメージを与えた時、
// 装備キャラクターは14/15/16/17/18の元素エネルギーを回復し、
// 12秒間Bountiful Sea's Sacred Wine効果を獲得する:
// HP上限がさらに14/16/18/20/22%増加し、
// Lunar Reactionダメージの会心ダメージが60/80/100/120/140%増加する。
// エネルギー回復効果は18秒毎に1回発動可能で、
// 装備キャラクターがフィールド外でも発動できる。

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		core: c,
		char: char,
	}
	r := p.Refine

	// 基本HP増加 (10/12/14/16/18%)
	baseHPBonus := 0.08 + 0.02*float64(r)
	val := make([]float64, attributes.EndStatType)
	val[attributes.HPP] = baseHPBonus

	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("nocturnes-curtain-call-base", -1),
		AffectedStat: attributes.HPP,
		Amount: func() ([]float64, bool) {
			return val, true
		},
	})

	// Lunar反応時のエネルギー回復とバフ
	energyAmt := []float64{14, 15, 16, 17, 18}
	buffHPBonus := []float64{0.14, 0.16, 0.18, 0.20, 0.22}
	buffCDBonus := []float64{0.60, 0.80, 1.00, 1.20, 1.40}

	const buffKey = "nocturnes-curtain-call-buff"
	const icdKey = "nocturnes-curtain-call-icd"
	buffDuration := 720 // 12s
	icdDuration := 1080 // 18s

	// Lunar反応イベントを購読
	lunarReactionTrigger := func(args ...interface{}) bool {
		// args[0]はreactableターゲット
		// 攻撃者がこのキャラクターかチェック
		atk, ok := args[1].(*combat.AttackEvent)
		if !ok {
			return false
		}
		if atk.Info.ActorIndex != char.Index {
			return false
		}

		// ICDをチェック
		if !char.StatusIsActive(icdKey) {
			// エネルギーを回復
			char.AddEnergy("nocturnes-curtain-call", energyAmt[r-1])
			c.Log.NewEvent("nocturnes curtain call energy recovery", glog.LogWeaponEvent, char.Index).
				Write("energy", energyAmt[r-1])
				// ICDを追加
			char.AddStatus(icdKey, icdDuration, true)
		}

		// バフを適用 (HP + Lunar反応ダメージの会心ダメージ)
		char.AddStatus(buffKey, buffDuration, true)

		return false
	}

	c.Events.Subscribe(event.OnLunarCharged, lunarReactionTrigger, fmt.Sprintf("nocturnes-curtain-call-lc-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnLunarBloom, lunarReactionTrigger, fmt.Sprintf("nocturnes-curtain-call-lb-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnLunarCrystallize, lunarReactionTrigger, fmt.Sprintf("nocturnes-curtain-call-lcrs-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk, ok := args[1].(*combat.AttackEvent)
		if !ok {
			return false
		}
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagLCDamage &&
			atk.Info.AttackTag != attacks.AttackTagLBDamage &&
			atk.Info.AttackTag != attacks.AttackTagLCrsDamage {
			return false
		}

		// ICDをチェック
		if !char.StatusIsActive(icdKey) {
			// エネルギーを回復
			char.AddEnergy("nocturnes-curtain-call", energyAmt[r-1])
			c.Log.NewEvent("nocturnes curtain call energy recovery", glog.LogWeaponEvent, char.Index).
				Write("energy", energyAmt[r-1])
				// ICDを追加
			char.AddStatus(icdKey, icdDuration, true)
		}

		// バフを適用 (HP + Lunar反応ダメージの会心ダメージ)
		char.AddStatus(buffKey, buffDuration, true)

		return false
	}, "nocturnes-curtain-call-enemy-dmg-"+char.Base.Key.String())

	// バフからのHPボーナス
	buffVal := make([]float64, attributes.EndStatType)
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase(fmt.Sprintf("%s-hp", buffKey), -1),
		AffectedStat: attributes.HPP,
		Amount: func() ([]float64, bool) {
			if !char.StatusIsActive(buffKey) {
				return nil, false
			}
			buffVal[attributes.HPP] = buffHPBonus[r-1]
			return buffVal, true
		},
	})

	// Lunar Reactionダメージの会心ダメージボーナス
	cdVal := make([]float64, attributes.EndStatType)
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase(fmt.Sprintf("%s-cd", buffKey), -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if !char.StatusIsActive(buffKey) {
				return nil, false
			}
			// Lunar Reactionダメージかチェック
			if atk.Info.AttackTag != attacks.AttackTagLCDamage &&
				atk.Info.AttackTag != attacks.AttackTagLBDamage &&
				atk.Info.AttackTag != attacks.AttackTagLCrsDamage {
				return nil, false
			}
			cdVal[attributes.CD] = buffCDBonus[r-1]
			return cdVal, true
		},
	})

	return w, nil
}
