package bloodsoakedruins

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.BloodsoakedRuins, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// 元素爆発使用後3.5秒間、装備キャラクターのLunar-Chargedダメージが36%(R1)/48%(R2)/60%(R3)/72%(R4)/84%(R5)増加する。
// さらに、Lunar-Charged反応発動後、装備キャラクターはRequiem of Ruinを獲得する: 会心ダメージが28%(R1)/35%(R2)/42%(R3)/49%(R4)/56%(R5)増加する（6秒間）。
// また、12/13/14/15/16の元素エネルギーを回復する。エネルギー回復は14秒毎に1回発動可能。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	const (
		lcDmgBuffKey = "bloodsoakedruins-lcdmg"
		cdBuffKey    = "bloodsoakedruins-cd"
		energyICDKey = "bloodsoakedruins-energy-icd"
		lcDmgBuffDur = 210 // 3.5s * 60 frames
		cdBuffDur    = 360 // 6s * 60 frames
		energyICD    = 840 // 14s * 60 frames
	)

	lcDmgBonus := 0.24 + float64(r)*0.12
	cdBonus := 0.21 + float64(r)*0.07
	energyRestore := float64(11 + r)

	// 効果1: 元素爆発使用後、Lunar-Chargedダメージを3.5秒間増加
	c.Events.Subscribe(event.OnBurst, func(args ...interface{}) bool {
		if c.Player.Active() != char.Index {
			return false
		}

		char.AddLCReactBonusMod(character.LCReactBonusMod{
			Base: modifier.NewBaseWithHitlag(lcDmgBuffKey, lcDmgBuffDur),
			Amount: func(ai combat.AttackInfo) (float64, bool) {
				return lcDmgBonus, false
			},
		})

		return false
	}, fmt.Sprintf("bloodsoakedruins-burst-%v", char.Base.Key.String()))

	// 効果2: Lunar-Charged反応発動後、会心ダメージバフを獲得しエネルギーを回復
	c.Events.Subscribe(event.OnLunarCharged, func(args ...interface{}) bool {
		ae := args[1].(*combat.AttackEvent)

		// キャラクターがLunar-Charged反応を発動したかチェック
		if ae.Info.ActorIndex != char.Index {
			return false
		}

		// 会心ダメージバフを適用
		mCD := make([]float64, attributes.EndStatType)
		mCD[attributes.CD] = cdBonus
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(cdBuffKey, cdBuffDur),
			AffectedStat: attributes.CD,
			Amount: func() ([]float64, bool) {
				return mCD, true
			},
		})

		// ICD付きでエネルギーを回復
		if !char.StatusIsActive(energyICDKey) {
			char.AddStatus(energyICDKey, energyICD, true)
			char.AddEnergy("bloodsoakedruins-energy", energyRestore)
		}

		return false
	}, fmt.Sprintf("bloodsoakedruins-lc-%v", char.Base.Key.String()))

	return w, nil
}
