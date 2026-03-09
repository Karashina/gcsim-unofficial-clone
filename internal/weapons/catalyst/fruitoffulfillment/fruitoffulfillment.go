package fruitoffulfillment

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
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
	core.RegisterWeaponFunc(keys.FruitOfFulfillment, NewWeapon)
}

type Weapon struct {
	Index  int
	core   *core.Core
	char   *character.CharWrapper
	stacks int
	// 装備者が必要チェックのスタック減少を確認
	stackLossTimer int
	lastStackGain  int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// 元素反応が発動した後、「盈欠」効果を獲得し、元素熟知が24/27/30/33/36増加するが攻撃力が5%減少する。
// 0.3秒毎にスタックを1獲得可能。最大5スタック。
// 元素反応が6秒間発動しなかった場合、スタックが1減少する。
// キャラクターがフィールドにいなくても発動可能。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		core: c,
		char: char,
	}
	r := p.Refine

	em := 21 + float64(r)*3
	atkLoss := -0.05

	w.stackLossTimer = 360 // 6s * 60

	const buffKey = "fruitoffulfillment"
	const icdKey = "fruitoffulfillment-icd"

	m := make([]float64, attributes.EndStatType)
	w.char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase(buffKey, -1),
		AffectedStat: attributes.NoStat,
		Amount: func() ([]float64, bool) {
			m[attributes.EM] = em * float64(w.stacks)
			m[attributes.ATKP] = atkLoss * float64(w.stacks)
			return m, true
		},
	})

	f := func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != w.char.Index {
			return false
		}
		if w.char.StatusIsActive(icdKey) {
			return false
		}
		w.char.AddStatus(icdKey, 18, true)

		w.stacks++
		if w.stacks > 5 {
			w.stacks = 5
		}

		w.lastStackGain = c.F
		w.char.QueueCharTask(w.checkStackLoss(c.F), w.stackLossTimer)

		w.core.Log.NewEvent("fruitoffulfillment gained stack", glog.LogWeaponEvent, w.char.Index).
			Write("stacks", w.stacks)

		return false
	}

	for i := event.ReactionEventStartDelim + 1; i < event.OnShatter; i++ {
		w.core.Events.Subscribe(i, f, fmt.Sprintf("fruitoffulfillment-%v", w.char.Base.Key.String()))
	}

	return w, nil
}

// スタック減少チェックのヘルパー関数
// スタック獲得毎に呼び出される
func (w *Weapon) checkStackLoss(src int) func() {
	return func() {
		if w.lastStackGain != src {
			w.core.Log.NewEvent("fruitoffulfillment stack loss check ignored, src diff", glog.LogWeaponEvent, w.char.Index).
				Write("src", src).
				Write("new src", w.lastStackGain)
			return
		}
		w.stacks--
		w.core.Log.NewEvent("fruitoffulfillment lost stack", glog.LogWeaponEvent, w.char.Index).
			Write("stacks", w.stacks).
			Write("last_stack_change", w.lastStackGain)

		// まだスタックがあれば再度キューに追加
		if w.stacks > 0 {
			w.char.QueueCharTask(w.checkStackLoss(src), w.stackLossTimer)
		}
	}
}
