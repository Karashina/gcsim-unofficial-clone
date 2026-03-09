package scholar

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

func init() {
	core.RegisterSetFunc(keys.Scholar, NewSet)
}

type Set struct {
	Index int
	Count int
}

func (s *Set) SetIndex(idx int) { s.Index = idx }
func (s *Set) GetCount() int    { return s.Count }
func (s *Set) Init() error      { return nil }

// 2セット効果: 元素チャージ効率 +20%
// 4セット効果: 元素粒子または元素オーブ獲得時、弓または法器を装備した全パーティメンバーにエネルギーを3付与。3秒に1回のみ。
func NewSet(c *core.Core, char *character.CharWrapper, count int, param map[string]int) (info.Set, error) {
	s := Set{Count: count}

	if count >= 2 {
		m := make([]float64, attributes.EndStatType)
		m[attributes.ER] = 0.20
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("scholar-2pc", -1),
			AffectedStat: attributes.ER,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}
	if count >= 4 {
		const icdKey = "scholar-4pc-icd"
		icd := 180
		// TODO: テスト未実施
		c.Events.Subscribe(event.OnParticleReceived, func(args ...interface{}) bool {
			if c.Player.Active() != char.Index {
				return false
			}
			if char.StatusIsActive(icdKey) {
				return false
			}
			char.AddStatus(icdKey, icd, true)

			for _, this := range c.Player.Chars() {
				// 弓と法器のみ対象
				if this.Weapon.Class == info.WeaponClassBow || this.Weapon.Class == info.WeaponClassCatalyst {
					this.AddEnergy("scholar-4pc", 3)
				}
			}

			return false
		}, fmt.Sprintf("scholar-4pc-%v", char.Base.Key.String()))
	}

	return &s, nil
}
