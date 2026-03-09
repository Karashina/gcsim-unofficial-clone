package jadefallssplendor

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/shield"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.JadefallsSplendor, NewWeapon)
}

type Weapon struct {
	Index int
	src   int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// 元素爆発使用またはシールド生成後3秒間、装備キャラクターはPrimordial Jade Regalia効果を獲得できる:
// 2.5秒毎に4.5/5/5.5/6/6.5のエネルギーを回復し、HP上限1,000毎に
// 対応する元素タイプの元素ダメージボーナスが0.3/0.5/0.7/0.9/1.1%増加する（最奇12/20/28/36/44%）。
// Primordial Jade Regaliaは装備キャラクターがフィールドにいなくても発動する。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	energy := 4 + float64(r)*0.5
	dmgMul := 0.001 + float64(r)*0.002
	dmgCap := 0.04 + float64(r)*0.08
	stat := attributes.EleToDmgP(char.Base.Element)

	const buffKey = "jadefall-buff"
	buffDuration := 3 * 60

	addBuff := func() {
		// エネルギー部分
		// バフが既にアクティブなら重複しないのでsrcが必要
		// baizhuのように6チックを得るために142を使用
		w.src = c.F
		char.QueueCharTask(w.addEnergy(c.F, energy, char), 142)

		// ダメージ部分
		finalDmg := char.MaxHP() * 0.001 * dmgMul
		if finalDmg > dmgCap {
			finalDmg = dmgCap
		}

		m := make([]float64, attributes.EndStatType)
		m[stat] = finalDmg
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(buffKey, buffDuration),
			AffectedStat: stat,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})
	}

	c.Events.Subscribe(event.OnBurst, func(args ...interface{}) bool {
		if c.Player.Active() != char.Index {
			return false
		}
		addBuff()
		return false
	}, fmt.Sprintf("jadefall-onburst-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnShielded, func(args ...interface{}) bool {
		shd := args[0].(shield.Shield)
		if shd.ShieldOwner() != char.Index {
			return false
		}
		addBuff()
		return false
	}, fmt.Sprintf("jadefall-onshielded-%v", char.Base.Key.String()))

	return w, nil
}

func (w *Weapon) addEnergy(src int, energy float64, char *character.CharWrapper) func() {
	return func() {
		if src != w.src {
			return
		}
		char.AddEnergy("jadefall-energy", energy)
	}
}
