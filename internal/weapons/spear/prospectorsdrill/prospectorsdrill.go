package prospectorsdrill

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
	symbolKey      = "prospectors-drill-symbol"
	symbolDuration = 30 * 60
	icdKey         = "prospectors-drill-icd"
	icdDuration    = 15 * 60
	buffKey        = "prospectors-drill"
	buffDuration   = 10 * 60
)

func init() {
	core.RegisterWeaponFunc(keys.ProspectorsDrill, NewWeapon)
}

type Weapon struct {
	stacks int
	Index  int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// 装備者が回復を受けた時、または他のキャラクターを回復した時、30秒間持続するUnityのシンボルを獲得する（最大3個）。
// 元素スキルまたは元素爆発使用時、全シンボルを消費して10秒間Struggle効果を獲得する。
// シンボル1個につき、攻撃力3/4/5/6/7%と全元素ダメージボーナス7/8.5/10/11.5/13%を獲得する。
// Struggle効果は15秒毎に1回発動可能。
// キャラクターがフィールドにいなくてもシンボルを獲得可能。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	// シンボルを獲得
	c.Events.Subscribe(event.OnHeal, func(args ...interface{}) bool {
		source := args[0].(*info.HealInfo)
		index := args[1].(int)
		amount := args[2].(float64)
		if source.Caller != char.Index && index != char.Index { // heal others and get healed including wielder
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
		c.Log.NewEvent("prospectors-drill adding stack", glog.LogWeaponEvent, char.Index).
			Write("stacks", w.stacks)
		char.AddStatus(symbolKey, symbolDuration, true)
		return false
	}, fmt.Sprintf("prospectors-drill-heal-%v", char.Base.Key.String()))

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
	key := fmt.Sprintf("prospectors-drill-struggle-%v", char.Base.Key.String())
	c.Events.Subscribe(event.OnBurst, buffFunc, key)
	c.Events.Subscribe(event.OnSkill, buffFunc, key)

	return w, nil
}
