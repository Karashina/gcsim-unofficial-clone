package rangegauge

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

const (
	symbolKey      = "range-gauge-symbol"
	symbolDuration = 30 * 60
	icdKey         = "range-gauge-icd"
	icdDuration    = 15 * 60
	buffKey        = "range-gauge"
	buffDuration   = 10 * 60
)

func init() {
	core.RegisterWeaponFunc(keys.RangeGauge, NewWeapon)
}

type Weapon struct {
	stacks int
	Index  int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// 装備者が回復を受けるまたは他者を回復した時、30秒間持続するUnityのシンボルを獲得する（最大3個）。
// 元素スキルまたは元素爆発使用時、全てのシンボルが消費されStruggle効果が10秒間付与される。
// シンボル1個消費毎に攻撃力3/4/5/6/7%と全元素ダメージバフ7/8.5/10/11.5/13%を獲得する。
// Struggle効果は15秒毎に1回のみ発動可能。
// キャラクターがフィールドにいなくてもシンボルを獲得可能。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	// シンボルを獲得
	c.Events.Subscribe(event.OnHeal, func(args ...interface{}) bool {
		source := args[0].(*info.HealInfo)
		index := args[1].(int)
		amount := args[2].(float64)
		if source.Caller != char.Index && index != char.Index { // 他者を回復または装備者が回復を受ける
			return false
		}
		if amount <= 0 {
			return false
		}

		if !char.StatusIsActive(symbolKey) {
			w.stacks = 0
		}
		if w.stacks < 3 {
			w.stacks++
		}
		c.Log.NewEvent("range-gauge adding stack", glog.LogWeaponEvent, char.Index).
			Write("stacks", w.stacks)
		char.AddStatus(symbolKey, symbolDuration, true)
		return false
	}, fmt.Sprintf("range-gauge-heal-%v", char.Base.Key.String()))

	// シンボルを消費
	baseEle := 0.055 + 0.015*float64(r)
	atk := 0.02 + 0.01*float64(r)
	m := make([]float64, attributes.EndStatType)
	buffFunc := func(args ...interface{}) bool {
		// シンボルがなければスキップ（ステータス非アクティブはシンボル == 0を意味する）
		if !char.StatusIsActive(symbolKey) {
			return false
		}
		// ICD中のトリガーはスキップ
		if char.StatusIsActive(icdKey) {
			return false
		}
		// シンボル削除前にアクティブ状態をチェック
		if c.Player.Active() != char.Index {
			return false
		}
		// ICDを追加
		char.AddStatus(icdKey, icdDuration, true)

		// シンボルを消費
		count := w.stacks
		char.DeleteStatus(symbolKey)
		w.stacks = 0

		// バフを追加
		m[attributes.ATKP] = atk * float64(count)
		ele := baseEle * float64(count)
		for i := attributes.PyroP; i <= attributes.DendroP; i++ {
			m[i] = ele
		}
		char.AddStatMod(character.StatMod{
			Base: modifier.NewBaseWithHitlag(buffKey, buffDuration),
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})

		return false
	}
	key := fmt.Sprintf("range-gauge-struggle-%v", char.Base.Key.String())
	c.Events.Subscribe(event.OnBurst, buffFunc, key)
	c.Events.Subscribe(event.OnSkill, buffFunc, key)

	return w, nil
}
