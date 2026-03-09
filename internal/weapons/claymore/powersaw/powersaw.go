package powersaw

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
	symbolKey      = "portable-power-saw-symbol"
	symbolDuration = 30 * 60
	icdKey         = "portable-power-saw-icd"
	icdDuration    = 15 * 60
	buffKey        = "portable-power-saw-em"
	buffDuration   = 10 * 60
)

func init() {
	core.RegisterWeaponFunc(keys.PortablePowerSaw, NewWeapon)
}

type Weapon struct {
	stacks int
	Index  int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// 装備者が回復するか他者を回復すると、30秒間持続するStoic's Symbolを獲得。最大3つ。
// 元素スキルまたは元素爆発を使用すると、全シンボルが消費されRousted効果が10秒間付与される。
// 消費したシンボル1つにつき元素熟知が40/50/60/70/80増加し、
// 効果発生の2秒後にシンボル1つにつき2/2.5/3/3.5/4のエネルギーが回復する。
// Roused効果は15秒に1回発動可能。
// キャラクターがフィールドにいなくてもシンボルを獲得可能。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	// シンボルを獲得
	c.Events.Subscribe(event.OnHeal, func(args ...interface{}) bool {
		source := args[0].(*info.HealInfo)
		index := args[1].(int)
		amount := args[2].(float64)
		if source.Caller != char.Index && index != char.Index { // 他者の回復と装備者自身の回復を含む
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
		c.Log.NewEvent("portable-power-saw adding stack", glog.LogWeaponEvent, char.Index).
			Write("stacks", w.stacks)
		char.AddStatus(symbolKey, symbolDuration, true)
		return false
	}, fmt.Sprintf("portable-power-saw-heal-%v", char.Base.Key.String()))

	// シンボルを消費
	em := 30 + 10*float64(r)
	refund := 1.5 + 0.5*float64(r)
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

		// 元素熟知バフを追加
		m[attributes.EM] = em * float64(count)
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(buffKey, buffDuration),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})

		// 2秒後にエネルギーを回復
		char.QueueCharTask(func() {
			char.AddEnergy("portable-power-saw-energy", refund*float64(count))
		}, 2*60)

		return false
	}
	key := fmt.Sprintf("portable-power-saw-roused-%v", char.Base.Key.String())
	c.Events.Subscribe(event.OnBurst, buffFunc, key)
	c.Events.Subscribe(event.OnSkill, buffFunc, key)

	return w, nil
}
