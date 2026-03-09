package alley

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
	core.RegisterWeaponFunc(keys.AlleyHunter, NewWeapon)
}

type Weapon struct {
	stacks           int
	active           bool
	lastActiveChange int
	Index            int
	core             *core.Core
	char             *character.CharWrapper
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }

// シミュレーション開始時にフィールド外ならスタック増加を開始
func (w *Weapon) Init() error {
	w.active = w.core.Player.Active() == w.char.Index
	if !w.active {
		w.core.Tasks.Add(w.incStack(w.char, w.core.F), 1)
	}
	w.lastActiveChange = w.core.F
	return nil
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	// 装備キャラクターがパーティにいるがフィールドにいない時、1秒毎にダメージが2%増加し、
	// 最大で20%まで。フィールドに4秒以上いると、前述のダメージバフが
	// 1秒毎に4%減少し、0になるまで続く。
	r := p.Refine

	// 最大10スタック
	w := Weapon{
		core: c,
		char: char,
	}
	w.stacks = p.Params["stacks"]
	if w.stacks > 10 {
		w.stacks = 10
	}
	dmg := 0.015 + float64(r)*0.005

	m := make([]float64, attributes.EndStatType)
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("alley-hunter", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			m[attributes.DmgP] = dmg * float64(w.stacks)
			return m, true
		},
	})

	key := fmt.Sprintf("alley-hunter-%v", char.Base.Key.String())

	c.Events.Subscribe(event.OnCharacterSwap, func(args ...interface{}) bool {
		prev := args[0].(int)
		next := args[1].(int)
		if next == char.Index {
			w.active = true
			w.lastActiveChange = c.F
			w.char.QueueCharTask(w.decStack(char, c.F), 240) // フィールドに4秒以上滞在したらスタック減少開始
		} else if prev == char.Index {
			w.active = false
			w.lastActiveChange = c.F
			c.Tasks.Add(w.incStack(char, c.F), 60)
		}
		return false
	}, key)

	return &w, nil
}

func (w *Weapon) decStack(c *character.CharWrapper, src int) func() {
	return func() {
		if w.active && w.stacks > 0 && src == w.lastActiveChange {
			w.stacks -= 2
			if w.stacks < 0 {
				w.stacks = 0
			}
			w.core.Log.NewEvent("Alley lost stack", glog.LogWeaponEvent, w.char.Index).
				Write("stacks:", w.stacks).
				Write("last_swap", w.lastActiveChange).
				Write("source", src)
			w.char.QueueCharTask(w.decStack(c, src), 60)
		}
	}
}

func (w *Weapon) incStack(c *character.CharWrapper, src int) func() {
	return func() {
		if !w.active && w.stacks < 10 && src == w.lastActiveChange {
			w.stacks++
			w.core.Log.NewEvent("Alley gained stack", glog.LogWeaponEvent, w.char.Index).
				Write("stacks:", w.stacks).
				Write("last_swap", w.lastActiveChange).
				Write("source", src)
			w.core.Tasks.Add(w.incStack(c, src), 60)
		}
	}
}
