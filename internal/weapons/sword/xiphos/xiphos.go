package xiphos

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.XiphosMoonlight, NewWeapon)
}

type Weapon struct {
	erBuff float64
	core   *core.Core
	char   *character.CharWrapper
	Index  int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }

func (w *Weapon) Init() error {
	w.updateStats()
	return nil
}

// 10秒毎に以下の効果が発動する: 装備キャラは元素熟知の
// 0.036%/0.045%/0.054%/0.063%/0.072%の元素チャージ効率ボーナスを12秒間得る。
// 近くのパーティメンバーはこのバフの30%を同じ時間得る。
// この武器の複数装備でバフを重複可能。
// キャラクターがフィールドにいなくても発動する。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		core: c,
		char: char,
	}
	r := p.Refine

	w.erBuff = 0.00027 + float64(r)*0.00009
	return w, nil
}

func (w *Weapon) updateStats() {
	val := make([]float64, attributes.EndStatType)
	val[attributes.ER] = w.erBuff * w.char.NonExtraStat(attributes.EM)
	w.char.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag("xiphos", 12*60),
		AffectedStat: attributes.ER,
		Extra:        true,
		Amount: func() ([]float64, bool) {
			return val, true
		},
	})

	valTeam := make([]float64, attributes.EndStatType)
	valTeam[attributes.ER] = val[attributes.ER] * 0.3
	for _, this := range w.core.Player.Chars() {
		if this.Index == w.char.Index {
			continue
		}

		this.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(fmt.Sprintf("xiphos-%v", w.char.Base.Key.String()), 12*60),
			AffectedStat: attributes.ER,
			Extra:        true,
			Amount: func() ([]float64, bool) {
				return valTeam, true
			},
		})
	}

	w.char.QueueCharTask(w.updateStats, 10*60)
}
