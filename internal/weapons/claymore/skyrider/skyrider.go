package skyrider

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
	core.RegisterWeaponFunc(keys.SkyriderGreatsword, NewWeapon)
}

type Weapon struct {
	Index  int
	stacks int
	buff   []float64
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	// 命中時、通常攻撃または重撃で攻撃力が6秒間6%増加。最大4スタック。0.5秒に1回発動可能。
	w := &Weapon{}
	r := p.Refine

	atkbuff := 0.05 + float64(r)*0.01
	w.buff = make([]float64, attributes.EndStatType)
	const icdKey = "skyrider-icd"

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if c.Player.Active() != char.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagNormal && atk.Info.AttackTag != attacks.AttackTagExtra {
			return false
		}
		if char.StatusIsActive(icdKey) {
			return false
		}
		// 全スタックの持続時間が切れた後に命中した場合、0にリセット
		if !char.StatModIsActive("skyrider") {
			w.stacks = 0
		}

		if w.stacks < 4 {
			w.stacks++
			// バフを更新
			w.buff[attributes.ATKP] = float64(w.stacks) * atkbuff
		}

		// バフタイマーを延長
		char.AddStatus(icdKey, 30, true)

		// 4スタック未満の間、命中毎にスタックを追加しバフを更新
		// 6秒間持続
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("skyrider", 360),
			AffectedStat: attributes.NoStat,
			Amount: func() ([]float64, bool) {
				return w.buff, true
			},
		})

		return false
	}, fmt.Sprintf("skyrider-greatsword-%v", char.Base.Key.String()))

	return w, nil
}
