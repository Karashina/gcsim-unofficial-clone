package toukaboushigure

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.ToukabouShigure, NewWeapon)
}

type Weapon struct {
	Index int
}

const (
	icdKey    = "toukaboushigure-icd"
	debuffKey = "toukaboushigure-cursed-parasol"
)

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// 攻撃が敵に命中した時、敵のう1体に10秒間「呪いの傘」を付与する。
// この効果は15秒毎に1回発動可能。「呪いの傘」持続中にその敵が倒された場合、CDが即リセットされる。
// この武器を装備したキャラは「呪いの傘」の影響を受けた敵に対してダメージが16/20/24/28/32%増加する。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = 0.12 + 0.04*float64(r)
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("toukaboushigure", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			e, ok := t.(*enemy.Enemy)
			if !ok {
				return nil, false
			}
			if !e.StatusIsActive(debuffKey) {
				return nil, false
			}
			return m, true
		},
	})

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		e, ok := args[0].(*enemy.Enemy)
		atk := args[1].(*combat.AttackEvent)
		if !ok {
			return false
		}
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if char.StatusIsActive(icdKey) {
			return false
		}
		char.AddStatus(icdKey, 15*60, true)
		e.AddStatus(debuffKey, 10*60, true)

		return false
	}, fmt.Sprintf("toukaboushigure-%v", char.Base.Key.String()))

	c.Events.Subscribe(event.OnTargetDied, func(args ...interface{}) bool {
		e, ok := args[0].(*enemy.Enemy)
		if !ok {
			return false
		}
		if !e.StatusIsActive(debuffKey) {
			return false
		}
		if !char.StatusIsActive(icdKey) {
			return false
		}
		char.DeleteStatus(icdKey)
		return false
	}, fmt.Sprintf("toukaboushigure-reset-%v", char.Base.Key.String()))

	return w, nil
}
