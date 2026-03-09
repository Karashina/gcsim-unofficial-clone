package dockhand

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
	symbolKey      = "dockhands-assistant-symbol"
	symbolDuration = 30 * 60
	icdKey         = "dockhands-assistant-icd"
	icdDuration    = 15 * 60
	buffKey        = "dockhands-assistant-em"
	buffDuration   = 10 * 60
)

func init() {
	core.RegisterWeaponFunc(keys.TheDockhandsAssistant, NewWeapon)
}

type Weapon struct {
	stacks int
	Index  int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// 装備者が回復を受けるか他者を回復させた時、「結束のシンボル」を得る。30秒持続、最大3個まで。
// 元素スキルまたは元素爆発使用時、全シンボルを消費し「覚醒」効果が10秒間付与される。
// シンボル1個につき元素熟知40/50/60/70/80を獲得、
// 効果発動2秒後にシンボル1個につき2/2.5/3/3.5/4のエネルギーが回復する。
// 覚醒効果は15秒毎に1回発動可能。
// キャラクターがフィールドにいなくてもシンボルを獲得可能。
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	// シンボルを獲得
	c.Events.Subscribe(event.OnHeal, func(args ...interface{}) bool {
		source := args[0].(*info.HealInfo)
		index := args[1].(int)
		amount := args[2].(float64)
		if source.Caller != char.Index && index != char.Index { // 装備者が回復した場合と回復を受けた場合の両方を含む
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
		c.Log.NewEvent("dockhands-assistant adding stack", glog.LogWeaponEvent, char.Index).
			Write("stacks", w.stacks)
		char.AddStatus(symbolKey, symbolDuration, true)
		return false
	}, fmt.Sprintf("dockhands-assistant-heal-%v", char.Base.Key.String()))

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
			Base:         modifier.NewBaseWithHitlag(buffKey, 10*60),
			AffectedStat: attributes.EM,
			Amount: func() ([]float64, bool) {
				return m, true
			},
		})

		// 2秒後にエネルギーを回復
		char.QueueCharTask(func() {
			char.AddEnergy("dockhands-assistant-energy", refund*float64(count))
		}, 2*60)

		return false
	}
	key := fmt.Sprintf("dockhands-assistant-roused-%v", char.Base.Key.String())
	c.Events.Subscribe(event.OnBurst, buffFunc, key)
	c.Events.Subscribe(event.OnSkill, buffFunc, key)

	return w, nil
}
