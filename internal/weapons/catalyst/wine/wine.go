package wine

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
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
	core.RegisterWeaponFunc(keys.WineAndSong, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	// 通常攻撃が敵に命中すると、ダッシュまたは代替ダッシュの
	// スタミナ消費が14%減少する（5秒間）。また、ダッシュまたは
	// 代替ダッシュを使用すると攻撃力が20%増加する（5秒間）。

	w := &Weapon{}
	r := p.Refine

	m := make([]float64, attributes.EndStatType)
	m[attributes.ATKP] = .15 + float64(r)*.05
	stamReduction := .12 + float64(r)*.02
	key := fmt.Sprintf("wineandsong-%v", char.Base.Key.String())
	c.Events.Subscribe(event.OnDash, func(args ...interface{}) bool {
		if c.Player.Active() != char.Index {
			return false
		}
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("wineandsong", 60*5),
			AffectedStat: attributes.NoStat,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
		return false
	}, key)

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)
		if c.Player.Active() != char.Index {
			return false
		}
		if ae.Info.ActorIndex != char.Index {
			return false
		}
		if ae.Info.AttackTag != attacks.AttackTagNormal {
			return false
		}

		c.Player.AddStamPercentMod(key, 300, func(a action.Action) (float64, bool) {
			if a == action.ActionDash {
				return -stamReduction, false
			}
			return 0, false
		})
		return false
	}, key)

	return w, nil
}
