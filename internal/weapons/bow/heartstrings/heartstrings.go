package heartstrings

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
	core.RegisterWeaponFunc(keys.SilvershowerHeartstrings, NewWeapon)
}

type Weapon struct {
	char       *character.CharWrapper
	core       *core.Core
	prevStacks int
	Index      int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

const (
	bondKey          = "heartstrings-bond"
	skillKey         = "heartstrings-skill"
	healingKey       = "heartstrings-healing"
	burstCRKey       = "heartstrings-cr"
	burstCRKeyCancel = "heartstrings-cr-cancel"
)

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	// 装備キャラクターはRemedy効果を獲得できる。
	// Remedyスタックを1/2/3所持時、HP上限が12%/24%/40%増加する。
	// 以下の条件を満たすとスタックを1つ獲得できる:
	// 元素スキル使用時に25秒間スタック1つ;
	// 命の契約の値が増加した時に25秒間スタック1つ;
	// 回復を行った時に20秒間スタック1つ。
	// 装備キャラクターがフィールドにいなくてもスタック獲得可能。
	// 各スタックの持続時間は独立してカウントされる。
	// さらに、3スタック時、元素爆発の会心率が28%増加する。
	// この効果はスタックが3未満になってから4秒後に解除される。
	w := &Weapon{
		char: char,
		core: c,
	}
	r := p.Refine

	hpStack := 0.09 + float64(r)*0.03
	hpMaxStack := 0.03 + float64(r)*0.01

	// Max HP増加
	mHP := make([]float64, attributes.EndStatType)
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("heartstrings", -1),
		AffectedStat: attributes.HPP,
		Amount: func() ([]float64, bool) {
			stacks := w.Stacks()
			mHP[attributes.HPP] = hpStack * float64(stacks)
			if stacks >= 3 {
				mHP[attributes.HPP] += hpMaxStack
			}
			return mHP, true
		},
	})

	// スキル使用時
	c.Events.Subscribe(event.OnSkill, func(args ...interface{}) bool {
		if c.Player.Active() != char.Index {
			return false
		}

		w.AddStack(skillKey, 25*60)
		return false
	}, fmt.Sprintf("heartstrings-%v", char.Base.Key.String()))

	// 命の契約獲得時
	c.Events.Subscribe(event.OnHPDebt, func(args ...interface{}) bool {
		index := args[0].(int)
		amount := args[1].(float64)

		if char.Index != index || amount > 0 {
			return false
		}

		w.AddStack(bondKey, 25*60)
		return false
	}, fmt.Sprintf("heartstrings-%v", char.Base.Key.String()))

	// 回復時
	c.Events.Subscribe(event.OnHeal, func(args ...interface{}) bool {
		src := args[0].(*info.HealInfo)

		if src.Caller != char.Index {
			return false
		}

		w.AddStack(healingKey, 20*60)
		return false
	}, fmt.Sprintf("heartstrings-%v", char.Base.Key.String()))

	// 3スタック時の元素爆発会心率バフ
	mCR := make([]float64, attributes.EndStatType)
	mCR[attributes.CR] = 0.21 + float64(r)*0.07
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase(burstCRKey, -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagElementalBurst {
				return nil, false
			}
			if w.Stacks() < 3 && !char.StatusIsActive(burstCRKeyCancel) {
				return nil, false
			}
			return mCR, true
		},
	})

	return w, nil
}

func (w *Weapon) AddStack(name string, duration int) {
	w.char.AddStatus(name, duration, true)
	w.char.QueueCharTask(func() {
		if !w.char.StatusIsActive(name) {
			w.OnUpdateStack()
		}
	}, duration+1)
	w.OnUpdateStack()
}

func (w *Weapon) Stacks() int {
	count := 0
	if w.char.StatusIsActive(skillKey) {
		count++
	}
	if w.char.StatusIsActive(bondKey) {
		count++
	}
	if w.char.StatusIsActive(healingKey) {
		count++
	}
	return count
}

func (w *Weapon) OnUpdateStack() {
	stacks := w.Stacks()
	w.core.Log.NewEvent("heartstrings update stacks", glog.LogWeaponEvent, w.char.Index).
		Write("stacks", stacks).
		Write("bol-stack", w.char.StatusIsActive(bondKey)).
		Write("skill-stack", w.char.StatusIsActive(skillKey)).
		Write("heal-stack", w.char.StatusIsActive(healingKey))

	if w.prevStacks == 3 && stacks < 3 {
		// 元素爆発会心率効果はスタックが3未満になってから4秒後に解除される。
		w.char.AddStatus(burstCRKeyCancel, 4*60, true)
	}
	w.prevStacks = stacks
}
