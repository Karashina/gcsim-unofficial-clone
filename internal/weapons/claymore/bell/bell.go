package bell

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/shield"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.TheBell, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	// ダメージを受けると最大HPの20%分を吸収するシールドを生成する。この
	// シールドは10秒間または破壊されるまで持続し、45秒に1回のみ発動可能。
	// シールドで保護されている間、キャラクターのダメージが12%増加する。
	const icdKey = "bell-icd"
	w := &Weapon{}
	r := p.Refine

	hp := 0.17 + float64(r)*0.03
	val := make([]float64, attributes.EndStatType)
	val[attributes.DmgP] = 0.09 + float64(r)*0.03

	c.Events.Subscribe(event.OnPlayerHPDrain, func(args ...interface{}) bool {
		di := args[0].(*info.DrainInfo)
		if !di.External {
			return false
		}
		if di.Amount <= 0 {
			return false
		}
		if char.StatusIsActive(icdKey) {
			return false
		}
		char.AddStatus(icdKey, 2700, true)

		c.Player.Shields.Add(&shield.Tmpl{
			ActorIndex: char.Index,
			Target:     char.Index,
			Src:        c.F,
			ShieldType: shield.Bell,
			Name:       "Bell",
			HP:         hp * char.MaxHP(),
			Ele:        attributes.NoElement,
			Expires:    c.F + 600,
		})
		return false
	}, fmt.Sprintf("bell-%v", char.Base.Key.String()))

	// シールド中の場合ダメージを追加
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("bell", -1),
		AffectedStat: attributes.NoStat,
		Amount: func() ([]float64, bool) {
			return val, c.Player.Shields.CharacterIsShielded(char.Index, c.Player.Active())
		},
	})

	return w, nil
}
