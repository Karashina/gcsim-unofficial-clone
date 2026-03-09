package player

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
)

type StamPercentModFunc func(action.Action) (float64, bool)

type stamPercentMod struct {
	Key    string
	Amount StamPercentModFunc
	Expiry int
	Event  glog.Event
}

func (h *Handler) StamPercentMod(a action.Action) float64 {
	n := 0
	amt := 0.0
	for _, mod := range h.stamPercentMods {
		if mod.Expiry > *h.F || mod.Expiry == -1 {
			x, done := mod.Amount(a)
			amt += x
			if !done {
				h.stamPercentMods[n] = mod
				n++
			}
		}
	}
	h.stamPercentMods = h.stamPercentMods[:n]
	return amt
}

func (h *Handler) StamPercentModIsActive(key string) bool {
	ind := -1
	for i, v := range h.stamPercentMods {
		if v.Key == key {
			ind = i
		}
	}
	// mod が存在しない
	if ind == -1 {
		return false
	}
	// 有効期限をチェック
	if h.stamPercentMods[ind].Expiry < *h.F && h.stamPercentMods[ind].Expiry > -1 {
		return false
	}
	return true
}

// TODO: ヒットラグの影響を受けるか不明？
func (h *Handler) AddStamPercentMod(key string, dur int, f StamPercentModFunc) {
	mod := stamPercentMod{
		Key:    key,
		Expiry: *h.F + dur,
		Amount: f,
	}
	ind := -1
	for i, v := range h.stamPercentMods {
		if v.Key == mod.Key {
			ind = i
		}
	}

	// 存在しなければ新規作成して追加
	if ind == -1 {
		mod.Event = h.Log.NewEvent("stam mod added", glog.LogStatusEvent, -1).
			Write("overwrite", false).
			Write("key", mod.Key).
			Write("expiry", mod.Expiry)
		mod.Event.SetEnded(mod.Expiry)
		h.stamPercentMods = append(h.stamPercentMods, mod)
		return
	}

	// そうでなければ有効期限切れでないかチェック
	if h.stamPercentMods[ind].Expiry > *h.F || h.stamPercentMods[ind].Expiry == -1 {
		h.Log.NewEvent(
			"stam mod refreshed", glog.LogStatusEvent, -1,
		).
			Write("overwrite", true).
			Write("key", mod.Key).
			Write("expiry", mod.Expiry)

		mod.Event = h.stamPercentMods[ind].Event
	} else {
		// 期限切れの場合はイベントを上書き
		mod.Event = h.Log.NewEvent("stam mod added", glog.LogStatusEvent, -1).
			Write("overwrite", false).
			Write("key", mod.Key).
			Write("expiry", mod.Expiry)
	}
	mod.Event.SetEnded(mod.Expiry)
	h.stamPercentMods[ind] = mod
}
