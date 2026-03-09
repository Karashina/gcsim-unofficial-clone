package wolffang

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
	core.RegisterWeaponFunc(keys.WolfFang, NewWeapon)
}

type Weapon struct {
	Index  int
	refine int
	c      *core.Core
	char   *character.CharWrapper
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// 元素スキルと元素爆発のダメージが16%増加する。元素スキルが敵に命中した時、
// 会心率が2%増加する。元素爆発が敵に命中した時、会心率が2%増加する。
// それぞれの効果は独立しており10秒持続、最大4スタック、0.1秒毎に1回発動可能。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		refine: p.Refine,
		c:      c,
		char:   char,
	}

	mFirst := make([]float64, attributes.EndStatType)
	mFirst[attributes.DmgP] = 0.12 + 0.04*float64(p.Refine)
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("wolf-fang", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			switch atk.Info.AttackTag {
			case attacks.AttackTagElementalArt:
			case attacks.AttackTagElementalArtHold:
			case attacks.AttackTagElementalBurst:
			default:
				return nil, false
			}
			return mFirst, true
		},
	})

	w.addEvent("wolf-fang-skill", attacks.AttackTagElementalArt, attacks.AttackTagElementalArtHold)
	w.addEvent("wolf-fang-burst", attacks.AttackTagElementalBurst)

	return w, nil
}

func (w *Weapon) addEvent(name string, tags ...attacks.AttackTag) {
	stacks := 0
	cr := 0.015 + 0.005*float64(w.refine)
	m := make([]float64, attributes.EndStatType)
	icd := name + "-icd"

	w.c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != w.char.Index {
			return false
		}
		if w.c.Player.Active() != w.char.Index {
			return false
		}
		if !requiredTag(atk.Info.AttackTag, tags...) {
			return false
		}
		if w.char.StatusIsActive(icd) {
			return false
		}
		w.char.AddStatus(icd, 0.1*60, true)

		if !w.char.StatusIsActive(name) {
			stacks = 0
		}
		if stacks < 4 {
			stacks++
		}

		w.char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBaseWithHitlag(name, 10*60),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				if !requiredTag(atk.Info.AttackTag, tags...) {
					return nil, false
				}
				m[attributes.CR] = cr * float64(stacks)
				return m, true
			},
		})
		return false
	}, fmt.Sprintf("%v-%v", name, w.char.Base.Key.String()))
}

func requiredTag(tag attacks.AttackTag, list ...attacks.AttackTag) bool {
	for _, value := range list {
		if tag == value {
			return true
		}
	}
	return false
}
