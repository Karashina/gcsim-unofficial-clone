package recurve

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
)

func init() {
	core.RegisterWeaponFunc(keys.RecurveBow, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// 敵を倒した時、HPが8/10/12/14/16%回復する。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	healPercentage := 0.06 + float64(r)*0.02
	c.Events.Subscribe(event.OnTargetDied, func(args ...interface{}) bool {
		_, ok := args[0].(*enemy.Enemy)
		// 敵でなければ無視
		if !ok {
			return false
		}
		atk := args[1].(*combat.AttackEvent)
		// 別のキャラクターが敵を倒した場合は発動しない
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		// フィールド外では発動しない
		if c.Player.Active() != char.Index {
			return false
		}
		// キャラクターを回復
		c.Player.Heal(info.HealInfo{
			Type:    info.HealTypePercent,
			Message: "Recurve Bow (Proc)",
			Src:     healPercentage,
		})
		return false
	}, fmt.Sprintf("recurvebow-%v", char.Base.Key.String()))

	return w, nil
}
