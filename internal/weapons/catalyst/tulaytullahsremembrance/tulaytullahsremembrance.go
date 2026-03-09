package tulaytullahsremembrance

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
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
	core.RegisterWeaponFunc(keys.TulaytullahsRemembrance, NewWeapon)
}

type Weapon struct {
	Index  int
	stacks int
	src    int
	core   *core.Core
}

const (
	icdKey    = "tulaytullahsremembrance-icd"
	atkSpdKey = "tulaytullahsremembrance-atkspd"
	buffKey   = "tulaytullahsremembrance-na-dmg"
)

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// 通常攻撃速度が10/12.5/15/17.5/20%増加する。
// 元素スキル発動後、14秒間毎秒通常攻撃ダメージが4.8/6/7.2/8.4/9.6%増加する。
// この持続時間中に通常攻撃が敵に命中すると、通常攻撃ダメージが9.6/12/14.4/16.8/19.2%増加する。
// この増加は0.3秒毎に1回発動可能。単一の効果持続時間あたりの最大通常攻撃ダメージ増加は48/60/72/84/96%。
// 装備者がフィールドを離れると効果が解除され、元素スキルを再度使用すると全ダメージバフがリセットされる。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		core: c,
	}
	r := p.Refine

	// 攻撃速度部分
	mAtkSpd := make([]float64, attributes.EndStatType)
	mAtkSpd[attributes.AtkSpd] = 0.075 + float64(r)*0.025
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase(atkSpdKey, -1),
		AffectedStat: attributes.NoStat,
		Amount: func() ([]float64, bool) {
			if c.Player.CurrentState() != action.NormalAttackState {
				return nil, false
			}
			return mAtkSpd, true
		},
	})

	// 通常攻撃ダメージ部分
	incDmg := 0.036 + float64(r)*0.012
	mDmg := make([]float64, attributes.EndStatType)
	c.Events.Subscribe(event.OnSkill, func(args ...interface{}) bool {
		if c.Player.Active() != char.Index {
			return false
		}

		// スキル使用時にスタックをリセット
		w.stacks = 0

		// スキル使用後14秒間、1秒毎にスタックを1獲得
		// 14秒のチェックは不要（最大スタック到達時に停止するため）
		w.src = c.F
		char.QueueCharTask(w.incStack(char, c.F), 60)

		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBaseWithHitlag(buffKey, 14*60),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				if atk.Info.AttackTag != attacks.AttackTagNormal {
					return nil, false
				}
				mDmg[attributes.DmgP] = incDmg * float64(w.stacks)
				return mDmg, true
			},
		})
		return false
	}, fmt.Sprintf("tulaytullahsremembrance-%v", char.Base.Key.String()))

	// 通常攻撃ダメージ時にスタックを2獲得、0.3秒ICD
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagNormal {
			return false
		}
		if char.StatusIsActive(icdKey) {
			return false
		}
		char.AddStatus(icdKey, 0.3*60, true)

		previous := w.stacks
		w.stacks += 2
		if w.stacks > 10 {
			w.stacks = 10
		}
		gain := w.stacks - previous
		if gain == 0 {
			return false
		}
		gainMsg := "2 stacks"
		if gain == 1 {
			gainMsg = "1 stack"
		}
		w.core.Log.NewEvent(fmt.Sprintf("Tulaytullah's Remembrance gained %v via normal attack", gainMsg), glog.LogWeaponEvent, char.Index).
			Write("stacks", w.stacks)
		return false
	}, fmt.Sprintf("tulaytullahsremembrance-ondmg-%v", char.Base.Key.String()))

	c.Events.Subscribe(event.OnCharacterSwap, func(args ...interface{}) bool {
		prev := args[0].(int)
		if prev != char.Index {
			return false
		}
		if !char.StatusIsActive(buffKey) {
			return false
		}
		// スタックを削除、incStackタスクを無効化、スワップ時にバフを削除
		w.stacks = 0
		w.src = -1
		char.DeleteStatus(buffKey)
		return false
	}, fmt.Sprintf("tulaytullahsremembrance-exit-%v", char.Base.Key.String()))

	return w, nil
}

func (w *Weapon) incStack(char *character.CharWrapper, src int) func() {
	return func() {
		if w.stacks > 9 {
			return
		}
		if src != w.src {
			return
		}
		w.stacks++
		w.core.Log.NewEvent("Tulaytullah's Remembrance gained stack via timer", glog.LogWeaponEvent, char.Index).
			Write("stacks", w.stacks).
			Write("weapon effect start", w.src).
			Write("source", src)
		char.QueueCharTask(w.incStack(char, src), 60)
	}
}
