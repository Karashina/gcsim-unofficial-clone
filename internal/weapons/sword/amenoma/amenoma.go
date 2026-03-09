package amenoma

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
)

func init() {
	core.RegisterWeaponFunc(keys.AmenomaKageuchi, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	seeds := []string{"amenoma-seed-0", "amenoma-seed-1", "amenoma-seed-2"}
	refund := 4.5 + 1.5*float64(r)
	const icdKey = "amenoma-icd"

	// TODO: 以前はpostskillだった。問題がないか確認が必要
	c.Events.Subscribe(event.OnSkill, func(args ...interface{}) bool {
		if c.Player.Active() != char.Index {
			return false
		}
		// 種を1つ追加
		if char.StatusIsActive(icdKey) {
			return false
		}
		// 最も古い種を見つけて上書き
		index := 0
		for i, s := range seeds {
			if char.StatusExpiry(s) < char.StatusExpiry(seeds[index]) {
				index = i
			}
		}
		char.AddStatus(seeds[index], 30*60, true)

		c.Log.NewEvent("amenoma proc'd", glog.LogWeaponEvent, char.Index).
			Write("index", index)

		char.AddStatus(icdKey, 300, true) // 5秒のICD

		return false
	}, fmt.Sprintf("amenoma-skill-%v", char.Base.Key.String()))

	// TODO: 以前はpostburstだった。問題がないか確認が必要
	c.Events.Subscribe(event.OnBurst, func(args ...interface{}) bool {
		if c.Player.Active() != char.Index {
			return false
		}
		count := 0
		for _, s := range seeds {
			if char.StatusIsActive(s) {
				count++
			}
			char.DeleteStatus(s)
		}
		if count == 0 {
			return false
		}
		// 2秒後にエネルギーを回復
		char.QueueCharTask(func() {
			char.AddEnergy("amenoma", refund*float64(count))
		}, 120+60) // 爆発アニメーション用に1秒追加したが、正確かは不明

		return false
	}, fmt.Sprintf("amenoma-burst-%v", char.Base.Key.String()))
	return w, nil
}
