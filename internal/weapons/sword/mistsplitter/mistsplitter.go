package mistsplitter

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
	core.RegisterWeaponFunc(keys.MistsplitterReforged, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

const (
	normalBuffKey = "mistsplitter-normal"
	burstBuffKey  = "mistsplitter-burst"
)

// 全元素ダメージボーナスが12%増加し、「霧切」の紋章の力を得る。
// スタック数が1/2/3の時、キャラクターの元素タイプの
// 元素ダメージボーナスが8/16/28%増加する。
// 以下の条件でそれぞれ1スタック獲得:
// 通常攻撃が元素ダメージを与えた時（5秒持続）、
// 元素爆発使用時（10秒持続）、
// エネルギーが100%未満の時（エネルギーが満タンになると消滅）。
// 各スタックの持続時間は独立して計算される。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	// 永続バフ
	m := make([]float64, attributes.EndStatType)
	base := 0.09 + float64(r)*0.03
	for i := attributes.PyroP; i <= attributes.DendroP; i++ {
		m[i] = base
	}

	// スタッキングバフ
	stack := 0.06 + float64(r)*0.02
	maxBonus := 0.03 + float64(r)*0.01
	bonus := attributes.EleToDmgP(char.Base.Element)

	// 通常攻撃がダメージを与えた時
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagNormal && atk.Info.AttackTag != attacks.AttackTagExtra {
			return false
		}
		if atk.Info.Element == attributes.Physical {
			return false
		}
		char.AddStatus(normalBuffKey, 300, true)
		return false
	}, fmt.Sprintf("mistsplitter-%v", char.Base.Key.String()))

	// 元素爆発使用時
	c.Events.Subscribe(event.OnBurst, func(args ...interface{}) bool {
		if c.Player.Active() != char.Index {
			return false
		}
		char.AddStatus(burstBuffKey, 600, true)
		return false
	}, fmt.Sprintf("mistsplitter-%v", char.Base.Key.String()))
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("mistsplitter", -1),
		AffectedStat: attributes.NoStat,
		Amount: func() ([]float64, bool) {
			count := 0
			if char.Energy < char.EnergyMax || char.EnergyMax == 0 {
				count++
			}
			if char.StatusIsActive(normalBuffKey) {
				count++
			}
			if char.StatusIsActive(burstBuffKey) {
				count++
			}
			dmg := float64(count) * stack
			if count >= 3 {
				dmg += maxBonus
			}
			m[bonus] = base + dmg
			return m, true
		},
	})

	return w, nil
}
