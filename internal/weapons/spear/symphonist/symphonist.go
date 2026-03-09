package symphonist

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const bufDur = 3 * 60
const buffKey = "symphonist-chanson-de-baies"

func init() {
	core.RegisterWeaponFunc(keys.SymphonistOfScents, NewWeapon)
}

type Weapon struct {
	Index int
	char  *character.CharWrapper
	c     *core.Core
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	// 攻撃力が12%増加する。装備キャラクターがフィールド外の時、攻撃力が
	// さらに12%増加する。回復を発動した時、装備キャラクターと回復対象は
	// 「Chanson de Baies」効果を獲得し、攻撃力が3秒間32%増加する。
	// この効果は装備キャラクターがフィールド外でも発動できる。
	w := &Weapon{
		char: char,
		c:    c,
	}
	r := p.Refine
	selfAtkP := 0.09 + float64(r)*0.03
	m := make([]float64, attributes.EndStatType)
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("symphonist-atkp", -1),
		AffectedStat: attributes.ATKP,
		Amount: func() ([]float64, bool) {
			m[attributes.ATKP] = selfAtkP
			if c.Player.Active() != char.Index {
				m[attributes.ATKP] += selfAtkP
			}
			return m, true
		},
	})

	buffOnHeal := make([]float64, attributes.EndStatType)
	buffOnHeal[attributes.ATKP] = 0.24 + float64(r)*0.08

	c.Events.Subscribe(event.OnHeal, func(args ...interface{}) bool {
		source := args[0].(*info.HealInfo)
		index := args[1].(int)

		if source.Caller != char.Index {
			return false
		}

		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(buffKey, bufDur),
			AffectedStat: attributes.ATKP,
			Amount: func() ([]float64, bool) {
				return buffOnHeal, true
			},
		})

		if index != char.Index {
			otherChar := c.Player.ByIndex(index)
			otherChar.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag(buffKey, bufDur),
				AffectedStat: attributes.ATKP,
				Amount: func() ([]float64, bool) {
					return buffOnHeal, true
				},
			})
		}
		return false
	}, fmt.Sprintf("symphonist-of-scents-%v", char.Base.Key.String()))

	return w, nil
}
