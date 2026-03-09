package ringofyaxche

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
	core.RegisterWeaponFunc(keys.RingOfYaxche, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// 元素スキル使用時、翡翠の冠効果を獲得する:
// HP上限1,000毎に、装備キャラクターの通常攻撃ダメージが
// 1%増加する（10秒間）。
// 通常攻撃ダメージの増加は最大32%まで。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	buffBy := 0.005 + 0.001*float64(r)
	maxBuff := 0.12 + 0.04*float64(r)

	c.Events.Subscribe(event.OnSkill, func(args ...interface{}) bool {
		if c.Player.Active() != char.Index {
			return false
		}

		buffAmt := min(maxBuff, char.MaxHP()*0.001*buffBy)
		m := make([]float64, attributes.EndStatType)
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase("ring-of-yaxche", 10*60),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				if atk.Info.AttackTag != attacks.AttackTagNormal {
					return nil, false
				}
				m[attributes.DmgP] = buffAmt
				return m, true
			},
		})
		return false
	}, fmt.Sprintf("ring-of-yaxche-%v", char.Base.Key.String()))

	return w, nil
}
