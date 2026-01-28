package flameforgedinsight

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.FlameForgedInsight, NewWeapon)
}

type Weapon struct {
	core  *core.Core
	char  *character.CharWrapper
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		core: c,
		char: char,
	}
	r := p.Refine

	f := func(args ...interface{}) bool {
		mEM := make([]float64, attributes.EndStatType)
		mEM[attributes.EM] = 45 + float64(r)*15
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("flameforgedinsight-em", 15*60),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				return mEM, true
			},
		})
		char.AddEnergy("flameforgedinsight-energy", 9+float64(r)*3)
		return false
	}

	c.Events.Subscribe(event.OnElectroCharged, f, "flameforgedinsight-ec")
	c.Events.Subscribe(event.OnLunarCharged, f, "flameforgedinsight-lc")
	c.Events.Subscribe(event.OnBloom, f, "flameforgedinsight-bloom")

	return w, nil
}
