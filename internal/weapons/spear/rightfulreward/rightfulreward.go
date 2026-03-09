package rightfulreward

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
)

func init() {
	core.RegisterWeaponFunc(keys.RightfulReward, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// 装備者が回復を受けた時、エネルギーを8/10/12/14/16回復する。
// この効果は10秒毎に1回発動可能で、キャラクターがフィールドにいなくても発動できる。

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	refund := 6 + float64(r)*2
	icd := 10 * 60
	const icdKey = "rightfulreward-icd"

	c.Events.Subscribe(event.OnHeal, func(args ...interface{}) bool {
		index := args[1].(int)
		if index != char.Index {
			return false
		}
		if char.StatusIsActive(icdKey) {
			return false
		}
		char.AddStatus(icdKey, icd, true) // 10秒ICD
		char.AddEnergy("rightfulreward", refund)

		return false
	}, fmt.Sprintf("rightfulreward-%v", char.Base.Key.String()))
	return w, nil
}
