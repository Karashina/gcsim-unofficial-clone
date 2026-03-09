package cinnabar

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
)

func init() {
	core.RegisterWeaponFunc(keys.CinnabarSpindle, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

const (
	icdKey      = "cinnabar-icd"
	durationKey = "cinnabar-buff-active"
)

// 元素スキルのダメージが防御力の40%分増加する。この効果は1.5秒毎に
// 1回のみ発動し、元素スキルがダメージを与えた0.1秒後にクリアされる。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	defPer := .3 + float64(r)*.1
	c.Events.Subscribe(event.OnEnemyHit, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagElementalArt && atk.Info.AttackTag != attacks.AttackTagElementalArtHold {
			return false
		}
		// ICD期間中は何もしない
		if char.StatusIsActive(icdKey) {
			return false
		}
		// 初回発動時は持続時間を設定しICDタスクをキューに追加
		if !char.StatusIsActive(durationKey) {
			// TODO: ICDは効果後に開始すると仮定
			char.QueueCharTask(func() {
				char.AddStatus(icdKey, 90, false) // ICDは1.5秒間
			}, 6) // ICDは6フレーム後に開始
			char.AddStatus(durationKey, 6, false)
		}
		damageAdd := char.TotalDef(false) * defPer
		atk.Info.FlatDmg += damageAdd

		c.Log.NewEvent("Cinnabar Spindle proc dmg add", glog.LogPreDamageMod, char.Index).
			Write("damage_added", damageAdd)
		return false
	}, fmt.Sprintf("cinnabar-%v", char.Base.Key.String()))

	return w, nil
}
