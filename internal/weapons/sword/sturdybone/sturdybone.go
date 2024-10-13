package sturdybone

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

func init() {
	core.RegisterWeaponFunc(keys.SturdyBone, NewWeapon)
}

type Weapon struct {
	Index int
	count int
}

const durationKey = "sturdybone-buff-active"

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	c.Events.Subscribe(event.OnDash, func(args ...interface{}) bool {
		c.Player.AddStamPercentMod("sturdybone-stam", -1, func(a action.Action) (float64, bool) {
			if a == action.ActionDash {
				return -0.15, false
			}
			return 0, false
		})
		if c.Player.Active() != char.Index {
			return false
		}
		if char.StatusIsActive(durationKey) {
			return false
		}
		w.count = 0
		char.AddStatus(durationKey, 7*60, false)
		return false
	}, "sturdybone-dash")

	c.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagNormal {
			return false
		}
		if w.count > 18 {
			char.DeleteStatus(durationKey)
			return false
		}
		if !char.StatusIsActive(durationKey) {
			return false
		}
		damageAdd := char.TotalAtk() * (0.12 + float64(r)*0.04)
		atk.Info.FlatDmg += damageAdd
		c.Log.NewEvent("Sturdy Bone proc dmg add", glog.LogPreDamageMod, char.Index).
			Write("damage_added", damageAdd)
		return false
	}, fmt.Sprintf("sturdybone-%v", char.Base.Key.String()))
	return w, nil
}
