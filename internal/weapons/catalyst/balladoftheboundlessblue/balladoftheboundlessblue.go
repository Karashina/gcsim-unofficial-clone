package balladoftheboundlessblue

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
	core.RegisterWeaponFunc(keys.BalladOfTheBoundlessBlue, NewWeapon)
}

type Weapon struct {
	Index int
}

// 通常攻撃または重撃が敵に命中した後6秒間、
// 通常攻撃ダメージが8/10/12/14/16%増加し、
// 重撃ダメージが6/7.5/8/10.5/12%増加する。最大3スタック。
// この効果は0.3秒毎に1回発動可能。

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	stacks := 0
	const buffIcd = "ballad-of-the-boundless-blue-icd"
	const buffKey = "ballad-of-the-boundless-blue-dmgp"
	na := make([]float64, attributes.EndStatType)
	ca := make([]float64, attributes.EndStatType)

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if c.Player.Active() != char.Index {
			return false
		}
		if char.StatusIsActive(buffIcd) {
			return false
		}

		if !char.StatModIsActive(buffKey) {
			stacks = 0
		}
		stacks++
		if stacks > 3 {
			stacks = 3
		}

		char.AddStatus(buffIcd, 0.3*60, true)
		switch atk.Info.AttackTag {
		case attacks.AttackTagNormal, attacks.AttackTagExtra:
			char.AddAttackMod(character.AttackMod{
				Base: modifier.NewBaseWithHitlag(buffKey, 6*60),
				Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
					switch atk.Info.AttackTag {
					case attacks.AttackTagNormal:
						na[attributes.DmgP] = (0.06 + 0.02*float64(r)) * float64(stacks)
						return na, true
					case attacks.AttackTagExtra:
						ca[attributes.DmgP] = (0.045 + 0.015*float64(r)) * float64(stacks)
						return ca, true
					default:
						return nil, false
					}
				},
			})
		}
		return false
	}, fmt.Sprintf("ballad-of-the-boundless-blue-%v", char.Base.Key.String()))

	return w, nil
}
