package wanderingevenstar

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
	core.RegisterWeaponFunc(keys.WanderingEvenstar, NewWeapon)
}

type Weapon struct {
	atkBuff float64
	core    *core.Core
	char    *character.CharWrapper
	Index   int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }

func (w *Weapon) Init() error {
	w.updateStats()
	return nil
}

// 10秒毎に以下の効果が発動する: 装備キャラクターは元素熟知の24%/30%/36%/42%/48%分の
// 攻撃力ボーナスを12秒間獲得し、近くのパーティメンバーはこのバフの30%を
// 同じ時間獲得する。複数のこの武器の効果は重複可能。
// キャラクターがフィールドにいなくても発動する。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		core: c,
		char: char,
	}
	r := p.Refine

	w.atkBuff = 0.18 + float64(r)*0.06
	return w, nil
}

func (w *Weapon) updateStats() {
	val := make([]float64, attributes.EndStatType)
	val[attributes.ATK] = w.atkBuff * w.char.NonExtraStat(attributes.EM)
	w.char.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag("wanderingevenstar", 12*60),
		AffectedStat: attributes.ATK,
		Extra:        true,
		Amount: func() ([]float64, bool) {
			return val, true
		},
	})

	valTeam := make([]float64, attributes.EndStatType)
	valTeam[attributes.ATK] = val[attributes.ATK] * 0.3
	for _, this := range w.core.Player.Chars() {
		if this.Index == w.char.Index {
			continue
		}

		this.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(fmt.Sprintf("wanderingevenstar-%v", w.char.Base.Key.String()), 12*60),
			AffectedStat: attributes.ATK,
			Extra:        true,
			Amount: func() ([]float64, bool) {
				return valTeam, true
			},
		})
	}

	w.char.QueueCharTask(w.updateStats, 10*60)
}
