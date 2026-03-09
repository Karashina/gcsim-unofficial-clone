package flowingpurity

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.FlowingPurity, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// 元素スキル使用時、全元素ダメージボーナスが15秒間8/10/12/14/16%増加し、
// 最大HPの24%相当の命の契約が付与される。この効果は10秒に1回発動可能。
// 命の契約が解除された時、1,000HPごとに全元素ダメージボーナスが2/2.5/3/3.5/4%増加、
// 最奇12/15/18/21/24%まで。この効果は15秒間持続する。

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	const icdKey = "flowingpurity-icd"
	const bondKey = "flowingpurity-bond"
	eledmg := 0.06 + float64(r)*0.02
	duration := 15 * 60
	icd := 10 * 60

	m := make([]float64, attributes.EndStatType)
	for i := attributes.PyroP; i <= attributes.DendroP; i++ {
		m[i] = eledmg
	}
	bond := make([]float64, attributes.EndStatType)
	hp := 0.24
	bondPercentage := 0.015 + float64(r)*0.005
	bondDMGPCap := 0.09 + float64(r)*0.03
	debt := 0.0

	c.Events.Subscribe(event.OnSkill, func(args ...interface{}) bool {
		if c.Player.Active() != char.Index {
			return false
		}
		if char.StatusIsActive(icdKey) {
			return false
		}
		char.AddStatus(icdKey, icd, true)

		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("flowingpurity-eledmg-boost", duration),
			AffectedStat: attributes.NoStat,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})

		if !char.StatusIsActive(bondKey) {
			debt = 0
		}
		char.AddStatus(bondKey, -1, true)

		char.ModifyHPDebtByRatio(hp)
		debt += char.CurrentHPDebt()
		bondDMGP := (debt / 1000) * bondPercentage // 契約解除後にバフを得るのでHP負債を使用
		if bondDMGP > bondDMGPCap {
			bondDMGP = bondDMGPCap
		}
		for i := attributes.PyroP; i <= attributes.DendroP; i++ {
			bond[i] = bondDMGP
		}

		return false
	}, fmt.Sprintf("flowingpurity-eledmg%v", char.Base.Key.String()))

	c.Events.Subscribe(event.OnHeal, func(args ...interface{}) bool {
		index := args[1].(int)
		if index != char.Index {
			return false
		}
		if char.CurrentHPDebt() > 0 {
			return false
		}
		if !char.StatusIsActive(bondKey) {
			return false
		}
		char.DeleteStatus(bondKey)

		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("flowingpurity-bond-eledmg-boost", duration),
			AffectedStat: attributes.NoStat,
			Amount: func() ([]float64, bool) {
				return bond, true
			},
		})
		return false
	}, fmt.Sprintf("flowingpurity-bondeledmg%v", char.Base.Key.String()))
	return w, nil
}
