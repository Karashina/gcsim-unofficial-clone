package calamity

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.CalamityQueller, NewWeapon)
}

type Weapon struct {
	Index        int
	stacks       int
	char         *character.CharWrapper
	c            *core.Core
	icd          int
	lastBuffGain int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }
func (w *Weapon) incStacks() func() {
	return func() {
		if w.stacks < 6 {
			w.stacks++
			if w.stacks != 6 {
				w.char.QueueCharTask(w.incStacks(), w.icd) // check again in 1s if stacks are not max
			}
		}
		w.c.Log.NewEvent("calamity gained stack", glog.LogWeaponEvent, w.char.Index).
			Write("stacks", w.stacks)
	}
}
func (w *Weapon) checkBuffExpiry(src int) func() {
	return func() {
		if w.lastBuffGain != src {
			w.c.Log.NewEvent("calamity buff expiry check ignored, src diff", glog.LogWeaponEvent, w.char.Index).
				Write("src", src).
				Write("new src", w.lastBuffGain)
			return
		}
		w.stacks = 0
		w.c.Log.NewEvent("calamity buff expired", glog.LogWeaponEvent, w.char.Index).
			Write("src", src).
			Write("lastBuffGain", w.lastBuffGain)
	}
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	// 全元素ダメージボーナス12%増加。元素スキル使用後、20秒間Consummationを獲得し、
	// 攻撃力が1秒あたり3.2%増加する。この攻撃力増加は最大6スタック。
	// 武器装備キャラクターがフィールドにいない時、Consummationの攻撃力増加は2倍になる。
	w := &Weapon{
		char: char,
		c:    c,
	}
	r := p.Refine

	// 固定元素ダメージボーナス
	dmg := .09 + float64(r)*.03
	m := make([]float64, attributes.EndStatType)
	m[attributes.PyroP] = dmg
	m[attributes.HydroP] = dmg
	m[attributes.CryoP] = dmg
	m[attributes.ElectroP] = dmg
	m[attributes.AnemoP] = dmg
	m[attributes.GeoP] = dmg
	m[attributes.DendroP] = dmg
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("calamity-dmg", -1),
		AffectedStat: attributes.NoStat,
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})

	const buffKey = "calamity-consummation"
	buffDuration := 1200 // 20s * 60
	w.icd = 60           // 1s * 60

	// スキル使用後のスタックあたりの攻撃力増加
	// フィールド外の場合ボーナスが2倍
	atkbonus := .024 + float64(r)*.008
	skillPressBonus := make([]float64, attributes.EndStatType)
	c.Events.Subscribe(event.OnSkill, func(args ...interface{}) bool {
		if c.Player.Active() != char.Index {
			return false
		}

		// Calamityバフ更新時にスタックがリセットされないと想定
		w.lastBuffGain = c.F
		char.QueueCharTask(w.checkBuffExpiry(c.F), buffDuration)
		char.QueueCharTask(w.incStacks(), w.icd)

		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(buffKey, buffDuration),
			AffectedStat: attributes.NoStat,
			Amount: func() ([]float64, bool) {
				atk := atkbonus * float64(w.stacks)
				if c.Player.Active() != char.Index {
					atk *= 2
				}
				skillPressBonus[attributes.ATKP] = atk

				return skillPressBonus, true
			},
		})

		return false
	}, fmt.Sprintf("calamity-queller-%v", char.Base.Key.String()))

	return w, nil
}
