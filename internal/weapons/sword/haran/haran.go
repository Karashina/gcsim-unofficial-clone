package haran

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.HaranGeppakuFutsu, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

const (
	icdKey             = "haran-icd"
	maxWavespikeStacks = 2
)

// 全元素ダメージボーナスが12%増加する。他のパーティメンバーが
// 元素スキルを使用すると、「波穂」スタックを1つ獲得する。最大2スタック。
// この効果は0.3秒毎に1回発動可能。この武器を装備したキャラが
// 元素スキルを使用すると、全「波穂」スタックが消費され「波乱」を獲得:
// スタック1つにつき通常攻撃ダメージが8秒間20%増加する。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	// 永続バフ
	m := make([]float64, attributes.EndStatType)
	base := 0.09 + float64(r)*0.03
	m[attributes.PyroP] = base
	m[attributes.HydroP] = base
	m[attributes.CryoP] = base
	m[attributes.ElectroP] = base
	m[attributes.AnemoP] = base
	m[attributes.GeoP] = base
	m[attributes.DendroP] = base
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("haran-ele-bonus", -1),
		AffectedStat: attributes.NoStat,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})

	wavespikeStacks := 0

	nonActiveFn := func() {
		// 0.3秒毎に1回
		if char.StatusIsActive(icdKey) {
			return
		}
		// スタックを追加
		wavespikeStacks++
		if wavespikeStacks > maxWavespikeStacks {
			wavespikeStacks = maxWavespikeStacks
		}
		c.Log.NewEvent("Haran gained a wavespike stack", glog.LogWeaponEvent, char.Index).Write("stack", wavespikeStacks)
		char.AddStatus(icdKey, 18, true)
	}

	val := make([]float64, attributes.EndStatType)
	activeFn := func() bool {
		val[attributes.DmgP] = (0.15 + float64(r)*0.05) * float64(wavespikeStacks)
		char.AddAttackMod(character.AttackMod{
			Base: modifier.NewBaseWithHitlag("ripping-upheaval", 480),
			Amount: func(atk *combat.AttackEvent, t combat.Target) ([]float64, bool) {
				if atk.Info.AttackTag != attacks.AttackTagNormal {
					return nil, false
				}
				return val, true
			},
		})
		wavespikeStacks = 0
		return false
	}

	// TODO: 以前はpostだった。問題がないか確認が必要
	c.Events.Subscribe(event.OnSkill, func(args ...interface{}) bool {
		if c.Player.Active() != char.Index {
			nonActiveFn()
			return false
		}
		if wavespikeStacks != 0 {
			return activeFn()
		}
		return false
	}, fmt.Sprintf("wavespike-%v", char.Base.Key.String()))

	return w, nil
}
