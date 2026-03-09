package nagamasa

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.KatsuragikiriNagamasa, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	// 元素スキルダメージが6%増加。元素スキルが敵に命中した後、
	// キャラクターはエネルギーを3失うが、6秒間2秒毎にエネルギーを3回復する。
	// この効果は10秒に1回発動可能。キャラクターが
	// フィールドにいなくても発動する。
	w := &Weapon{}
	r := p.Refine

	m := make([]float64, attributes.EndStatType)
	base := 0.045 + float64(r)*0.015
	regen := 2.5 + float64(r)*0.5

	m[attributes.DmgP] = base
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("nagamasa-skill-dmg-buff", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag == attacks.AttackTagElementalArt || atk.Info.AttackTag == attacks.AttackTagElementalArtHold {
				return m, true
			}
			return nil, false
		},
	})

	const icdKey = "nagamasa-icd"
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagElementalArt && atk.Info.AttackTag != attacks.AttackTagElementalArtHold {
			return false
		}
		if char.StatusIsActive(icdKey) {
			return false
		}
		char.AddStatus(icdKey, 600, true)
		char.AddEnergy("nagamasa", -3)
		for i := 120; i <= 360; i += 120 {
			// ヒットラグ用にキャラキューを使用
			char.QueueCharTask(func() {
				char.AddEnergy("nagamasa", regen)
			}, i)
		}
		return false
	}, fmt.Sprintf("nagamasa-%v", char.Base.Key.String()))

	return w, nil
}
