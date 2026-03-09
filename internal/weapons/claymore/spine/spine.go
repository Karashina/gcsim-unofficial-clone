package spine

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
	core.RegisterWeaponFunc(keys.SerpentSpine, NewWeapon)
}

type Weapon struct {
	Index  int
	char   *character.CharWrapper
	c      *core.Core
	stacks int
	dmg    float64
	buff   []float64
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }
func (w *Weapon) stackCheck() func() {
	return func() {
		// フィールドにいてスタックが5未満の場合、スタックを追加
		if w.char.Index == w.c.Player.Active() {
			if w.stacks < 5 {
				w.stacks++
				w.updateBuff()
			}
		}
		w.char.QueueCharTask(w.stackCheck(), 240) // 4秒後に再チェック
	}
}
func (w *Weapon) updateBuff() {
	w.buff[attributes.DmgP] = float64(w.stacks) * w.dmg
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	// キャラクターがフィールドにいる間4秒毎に、ダメージが6%増加し
	// 被ダメージも3%増加する。この効果は最大5スタックで、キャラクターが
	// フィールドを離れてもリセットされないが、ダメージを受けると
	// 1スタック減少する。
	w := &Weapon{
		char: char,
		c:    c,
		buff: make([]float64, attributes.EndStatType),
	}
	r := p.Refine

	// 被ダメージ/スタック減少には1秒の内部クールダウンがある
	// それ以外は4秒毎にチェックし、フィールドにいればスタックを追加する
	// ゲーム内で確認済み: 最初の剣追加後、交代して戻ると
	// フィールド外にいた時間が約1秒でも次の剣アニメーションは4秒後

	w.dmg = 0.05 + float64(r)*.01
	// 初期値を設定
	w.stacks = p.Params["stacks"]
	c.Log.NewEvent(
		"serpent spine stack check", glog.LogWeaponEvent, char.Index,
	).
		Write("params", p.Params)

	if w.stacks > 5 {
		w.stacks = 5
	}
	w.updateBuff()

	// スタック増加チェック用のティッカーを開始
	char.QueueCharTask(w.stackCheck(), 240)

	// ダメージチェック用のイベントフックを追加（1秒のICDあり）
	//TODO: 被ダメージ3%増加は未実装
	const icdKey = "spine-dmgtaken-icd"
	icd := 60
	c.Events.Subscribe(event.OnPlayerHPDrain, func(args ...interface{}) bool {
		di := args[0].(*info.DrainInfo)
		if !di.External {
			return false
		}
		if di.Amount <= 0 {
			return false
		}
		if c.Player.Active() != char.Index {
			return false
		}
		if char.StatusIsActive(icdKey) {
			return false
		}
		char.AddStatus(icdKey, icd, true)
		if w.stacks > 0 {
			w.stacks--
			w.updateBuff()
		}
		return false
	}, fmt.Sprintf("spine-%v", char.Base.Key.String()))

	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("spine", -1),
		AffectedStat: attributes.NoStat,
		Amount: func() ([]float64, bool) {
			return w.buff, w.stacks > 0
		},
	})

	return w, nil
}
