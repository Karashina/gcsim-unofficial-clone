package firstgreatmagic

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.TheFirstGreatMagic, NewWeapon)
}

type Weapon struct {
	Index            int
	core             *core.Core
	char             *character.CharWrapper
	atkStackVal      float64
	sameElement      int
	differentElement int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error {
	// GimmickとTheatricsのスタックを計算
	for _, x := range w.core.Player.Chars() {
		if x.Base.Element == w.char.Base.Element { // 装備者を含む
			w.sameElement++
			continue
		}
		w.differentElement++
	}
	// バフ値計算のため元素数を上限クリップ
	if w.sameElement > 3 {
		w.sameElement = 3
	}
	if w.differentElement > 3 {
		w.differentElement = 3
	}

	// Gimmickバフ
	mAtk := make([]float64, attributes.EndStatType)
	mAtk[attributes.ATKP] = w.atkStackVal * float64(w.sameElement)
	w.char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("thefirstgreatmagic-atk", -1),
		AffectedStat: attributes.ATKP,
		Amount: func() ([]float64, bool) {
			return mAtk, true
		},
	})

	// Theatricsバフ
	// TODO: 移動速度は未実装

	return nil
}

// 重撃のダメージが16/20/24/28/32%増加する。
// 装備者と同じ元素タイプのパーティメンバー（装備者含む）1人につきGimmickスタックを1獲得。
// 装備者と異なる元素タイプのパーティメンバー1人につきTheatricsスタックを1獲得。
// Gimmickスタックが1/2/3以上の時、攻撃力が16%/32%/48% / 20%/40%/60% / 24%/48%/72% / 28%/56%/84% / 32%/64%/96%増加。
// Theatricsスタックが1/2/3以上の時、移動速度が4%/7%/10% / 6%/9%/12% / 8%/11%/14% / 10%/13%/16% / 12%/15%/18%増加。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		core: c,
		char: char,
	}
	r := p.Refine

	// 重撃バフ
	mDmg := make([]float64, attributes.EndStatType)
	mDmg[attributes.DmgP] = (0.12 + float64(r)*0.04)
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("thefirstgreatmagic-dmg%", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag != attacks.AttackTagExtra {
				return nil, false
			}
			return mDmg, true
		},
	})

	w.atkStackVal = (0.12 + float64(r)*0.04)

	return w, nil
}
