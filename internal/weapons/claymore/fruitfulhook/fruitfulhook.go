package fruitfulhook

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
	core.RegisterWeaponFunc(keys.FruitfulHook, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// 落下攻撃の会心率が16%増加。
// 落下攻撃が敵に命中した後、通常攻撃、重撃、落下攻撃のダメージが10秒間16%増加する。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	// 落下攻撃の会心率を増加
	mCR := make([]float64, attributes.EndStatType)
	mCR[attributes.CR] = 0.12 + 0.04*float64(r)
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("fruitful-hook-cr", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag == attacks.AttackTagPlunge {
				return mCR, true
			}
			return nil, false
		},
	})

	// 落下攻撃が敵に命中した後、通常攻撃、重撃、落下攻撃のダメージが10秒間増加
	mDMG := make([]float64, attributes.EndStatType)
	mDMG[attributes.DmgP] = 0.12 + 0.04*float64(r)
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagPlunge {
			return false
		}
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBaseWithHitlag("fruitful-hook-dmg%", 10*60),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				switch atk.Info.AttackTag {
				case attacks.AttackTagNormal:
				case attacks.AttackTagExtra:
				case attacks.AttackTagPlunge:
				default:
					return nil, false
				}
				return mDMG, true
			},
		})

		return false
	}, fmt.Sprintf("fruitful-hook-%v", char.Base.Key.String()))

	return w, nil
}
