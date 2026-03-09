package fleuvecendreferryman

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
	core.RegisterWeaponFunc(keys.FleuveCendreFerryman, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// 元素スキルの会心率が8/10/12/14/16%増加する。また、元素スキル使用後5秒間、元素チャージ効率が16/20/24/28/32%増加する。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	// スキル会心率
	mCR := make([]float64, attributes.EndStatType)
	mCR[attributes.CR] = 0.06 + 0.02*float64(r)
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("fleuvecendreferryman-cr", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag == attacks.AttackTagElementalArt || atk.Info.AttackTag == attacks.AttackTagElementalArtHold {
				return mCR, true
			}
			return nil, false
		},
	})

	// スキル後に元素チャージ効率アップ
	mER := make([]float64, attributes.EndStatType)
	mER[attributes.ER] = 0.12 + 0.04*float64(r)
	c.Events.Subscribe(event.OnSkill, func(args ...interface{}) bool {
		if c.Player.Active() != char.Index {
			return false
		}
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("fleuvecendreferryman-er", 5*60),
			AffectedStat: attributes.ER,
			Amount: func() ([]float64, bool) {
				return mER, true
			},
		})
		return false
	}, fmt.Sprintf("fleuvecendreferryman-%v", char.Base.Key.String()))

	return w, nil
}
