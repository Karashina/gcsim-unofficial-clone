package whiteblind

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
	core.RegisterWeaponFunc(keys.Whiteblind, NewWeapon)
}

type Weapon struct {
	Index  int
	stacks int
	buff   []float64
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	// 命中時、通常攻撃または重撃で攻撃力と防御力が6秒間6%増加。最大4
	// スタック。この効果は0.5秒に1回のみ発動。
	w := &Weapon{}
	r := p.Refine

	w.buff = make([]float64, attributes.EndStatType)
	amt := 0.045 + float64(r)*0.015
	const icdKey = "whiteblind-icd"

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
		if char.StatModIsActive(icdKey) {
			return false
		}
		if !char.StatModIsActive("whiteblind") {
			w.stacks = 0
		}

		char.AddStatus(icdKey, 30, true)

		if w.stacks < 4 {
			w.stacks++
			// バフを更新
			w.buff[attributes.ATKP] = amt * float64(w.stacks)
			w.buff[attributes.DEFP] = amt * float64(w.stacks)
		}

		// 修飾子を更新
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("whiteblind", 360),
			AffectedStat: attributes.NoStat,
			Amount: func() ([]float64, bool) {
				return w.buff, true
			},
		})

		return false
	}, fmt.Sprintf("whiteblind-%v", char.Base.Key.String()))

	return w, nil
}
