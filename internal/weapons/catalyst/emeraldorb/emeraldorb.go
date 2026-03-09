package emeraldorb

import (
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
	core.RegisterWeaponFunc(keys.EmeraldOrb, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// 蒸発、感電、凍結、または水元素付きの拡散反応を起こした時、攻撃力が20/25/30/35/40%増加する（12秒間）。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	m := make([]float64, attributes.EndStatType)
	m[attributes.ATKP] = 0.15 + float64(r)*0.05

	addBuff := func(args ...interface{}) bool {
		if _, ok := args[0].(*enemy.Enemy); !ok {
			return false
		}

		atk := args[1].(*combat.AttackEvent)
		// 武器装備者からのダメージでなければ発動しない
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		// フィールド外では発動しない
		if c.Player.Active() != char.Index {
			return false
		}

		// バフを追加
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("emeraldorb", 720),
			AffectedStat: attributes.NoStat,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})

		return false
	}

	subKey := "emeraldorb-" + char.Base.Key.String()

	c.Events.Subscribe(event.OnVaporize, addBuff, subKey)
	c.Events.Subscribe(event.OnElectroCharged, addBuff, subKey)
	c.Events.Subscribe(event.OnFrozen, addBuff, subKey)
	c.Events.Subscribe(event.OnSwirlHydro, addBuff, subKey)

	return w, nil
}
