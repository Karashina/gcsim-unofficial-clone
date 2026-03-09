package sealord

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
	core.RegisterWeaponFunc(keys.LuxuriousSeaLord, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	// 元素爆発のダメージが12%増加。元素爆発が敵に命中すると、
	// 100%の確率で攻撃力の100%分の範囲ダメージを与えるマグロの群れを召喚する。
	// この効果は15秒に1回発動可能。
	w := &Weapon{}
	r := p.Refine

	// 永続の元素爆発ダメージ増加
	burstDmgIncrease := .09 + float64(r)*0.03
	val := make([]float64, attributes.EndStatType)
	val[attributes.DmgP] = burstDmgIncrease
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("luxurious-sea-lord", -1),
		Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
			if atk.Info.AttackTag == attacks.AttackTagElementalBurst {
				return val, true
			}
			return nil, false
		},
	})

	tunaDmg := .75 + float64(r)*0.25
	const icdKey = "sealord-icd"
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...interface{}) bool {
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != char.Index {
			return false
		}
		if c.Player.Active() != char.Index {
			return false
		}
		if char.StatusIsActive(icdKey) {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagElementalBurst {
			return false
		}
		char.AddStatus(icdKey, 900, true)
		ai := combat.AttackInfo{
			ActorIndex: char.Index,
			Abil:       "Luxurious Sea-Lord Proc",
			AttackTag:  attacks.AttackTagWeaponSkill,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Physical,
			Durability: 100,
			Mult:       tunaDmg,
		}
		trg := args[0].(combat.Target)
		c.QueueAttack(ai, combat.NewCircleHitOnTarget(trg, nil, 3), 0, 1)

		return false
	}, fmt.Sprintf("sealord-%v", char.Base.Key.String()))
	return w, nil
}
