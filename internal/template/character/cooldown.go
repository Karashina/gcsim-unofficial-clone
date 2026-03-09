// Package cooldown は SetCD, SetCDWithDelay, ResetActionCooldown, ReduceActionCooldown, ActionReady のデフォルト実装を提供する
package character

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
)

// SetCD は2つのパラメータを受け取る:
//   - a action.Action: クールダウンをトリガーするアクションタイプ
//   - dur: クールダウンが持続するフレーム数
//
// AvailableCDCharges[a] > 0 であることが前提（そうでなければアクションは許可されないはず）
//
// SetCD はクールダウン時間をキューに追加する。これは複数チャージがある場合、
// ゲームが最初のチャージの回復を完了してから次のチャージの完全なクールダウンを
// 開始するためである。
//
// クールダウンが初めてキューに追加されると、キューワーカーが開始される。このワーカーは
// キューの最初のアイテムに指定されたクールダウン時間後にチェックし、キューのクールダウンが
// 変更されていなければチャージ数を1増やし、次のアイテムのためにリスケジュールする。
//
// ReduceActionCooldown や ResetActionCooldown によりキューのクールダウンが調整される
// ことがある。この場合、最初のワーカーは間違ったタイミングでチェックバックしてしまう。
// これを防ぐため、cdQueueWorkerStartedAt[a] でワーカーの開始フレームを追跡する。
// ReduceActionCooldown や ResetActionCooldown が呼ばれると新しいワーカーを開始し、
// cdQueueWorkerStartedAt[a] を更新する。これにより古いワーカーはこの値を確認して
// 自身の開始フレームと一致しなければ正常に終了できる。
func (c *Character) SetCD(a action.Action, dur int) {
	// CDの設定は回復キューにCDを追加するだけ
	// 正しい時間が追加されるようにまずクールダウン短縮をチェックする
	modified := c.CDReduction(a, dur)
	// 現在のアクションと時間をキューに追加
	c.cdQueue[a] = append(c.cdQueue[a], modified)
	// 追加前にキューが空だった場合、クールダウンキューワーカーを開始する
	if len(c.cdQueue[a]) == 1 {
		c.startCooldownQueueWorker(a)
	}
	// スタックカウントから1を減らす
	c.AvailableCDCharge[a]--
	if c.AvailableCDCharge[a] < 0 {
		panic("unexpected charges less than 0")
	}
	c.Core.Log.NewEventBuildMsg(glog.LogCooldownEvent, c.Index, a.String(), " cooldown triggered").
		Write("type", a.String()).
		Write("expiry", c.Cooldown(a)).
		Write("original_cd", dur).
		Write("modified_cd_by_cdr", modified).
		Write("charges_remain", c.AvailableCDCharge[a]).
		Write("cooldown_queue", c.cdQueueString(a))
}

func (c *Character) cdQueueString(a action.Action) string {
	var sb strings.Builder
	for i, v := range c.cdQueue[a] {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(strconv.Itoa(v))
	}
	return sb.String()
}

func (c *Character) SetNumCharges(a action.Action, num int) {
	c.additionalCDCharge[a] = num - 1
	c.AvailableCDCharge[a] = num
}

func (c *Character) Charges(a action.Action) int {
	return c.AvailableCDCharge[a]
}

func (c *Character) SetCDWithDelay(a action.Action, dur, delay int) {
	if delay == 0 {
		c.SetCD(a, dur)
		return
	}
	c.Core.Tasks.Add(func() { c.SetCD(a, dur) }, delay)
}

func (c *Character) Cooldown(a action.Action) int {
	// 残りクールダウン = src + キューの最初のアイテム - 現在のフレーム
	if c.AvailableCDCharge[a] > 0 {
		return 0
	}
	// そうでなければキューを確認; ゼロなら準備完了
	if len(c.cdQueue) == 0 {
		// panic("queue length is somehow 0??")
		return 0
	}
	return c.cdQueueWorkerStartedAt[a] + c.cdQueue[a][0] - c.Core.F
}

func (c *Character) ResetActionCooldown(a action.Action) {
	// スタックが既に最大なら何もしない
	if c.AvailableCDCharge[a] == 1+c.additionalCDCharge[a] {
		return
	}
	// log.Printf("resetting; frame %v, queue %v\n", c.F, c.cdQueue[a])
	// スタックを追加してキューをポップ
	c.AvailableCDCharge[a]++
	c.Tags["skill_charge"]++
	c.cdQueue[a] = c.cdQueue[a][1:]
	// ワーカー時間をリセット
	c.cdQueueWorkerStartedAt[a] = c.Core.F
	c.cdCurrentQueueWorker[a] = nil
	c.Core.Log.NewEventBuildMsg(glog.LogCooldownEvent, c.Index, a.String(), " cooldown forcefully reset").
		Write("type", a.String()).
		Write("charges_remain", c.AvailableCDCharge[a]).
		Write("cooldown_queue", c.cdQueueString(a))
	// キューに残りのCDがあるか確認
	if len(c.cdQueue) > 0 {
		c.startCooldownQueueWorker(a)
	}
}

func (c *Character) ReduceActionCooldown(a action.Action, v int) {
	// スタックが既に最大なら何もしない
	if c.AvailableCDCharge[a] == 1+c.additionalCDCharge[a] {
		return
	}
	// 短縮量が残り時間を超えるか確認。超える場合はCDリセットを呼ぶ
	remain := c.cdQueueWorkerStartedAt[a] + c.cdQueue[a][0] - c.Core.F
	// log.Printf("hello reducing; reduction %v, remaining %v, frame %v, old queue %v\n", v, remain, c.F, c.cdQueue[a])
	if v >= remain {
		c.ResetActionCooldown(a)
		return
	}
	// 残り時間を短縮してキューを再開
	c.cdQueue[a][0] = remain - v
	c.Core.Log.NewEventBuildMsg(glog.LogCooldownEvent, c.Index, a.String(), " cooldown forcefully reduced").
		Write("type", a.String()).
		Write("expiry", c.Cooldown(a)).
		Write("charges_remain", c.AvailableCDCharge).
		Write("cooldown_queue", c.cdQueueString(a))
	c.startCooldownQueueWorker(a)
	// log.Printf("started: %v, new queue: %v, worker frame: %v\n", c.cdQueueWorkerStartedAt[a], c.cdQueue[a], c.cdQueueWorkerStartedAt[a])
}

func (c *Character) startCooldownQueueWorker(a action.Action) {
	// アクション a のキューの長さを確認し、空なら開始するものがない
	if len(c.cdQueue[a]) == 0 {
		return
	}

	// このワーカーの開始時間を設定
	c.cdQueueWorkerStartedAt[a] = c.Core.F
	var src *func()

	worker := func() {
		// srcが変更されていたら何もしない
		if src != c.cdCurrentQueueWorker[a] {
			// c.Log.Debugw("src changed",  "src", src, "new", c.cdQueueWorkerStartedAt[a])
			return
		}
		// log.Printf("cd worker triggered, started; %v, queue: %v\n", c.cdQueueWorkerStartedAt[a], c.cdQueue[a])
		// キューが空でないことを確認
		if len(c.cdQueue[a]) == 0 {
			// これは発生しないはず
			panic(fmt.Sprintf(
				"queue is empty? index :%v, frame : %v, worker src: %v, started: %v",
				c.Index,
				c.Core.F,
				src,
				c.cdQueueWorkerStartedAt[a],
			))
			// return
		}
		// スタックを追加してキューの最初のアイテムをポップ
		c.AvailableCDCharge[a]++
		c.Tags["skill_charge"]++
		c.cdQueue[a] = c.cdQueue[a][1:]

		// c.Log.Debugw("stack restored",  "avail", c.availableCDCharge[a], "queue", c.cdQueue)

		if c.AvailableCDCharge[a] > 1+c.additionalCDCharge[a] {
			// 安全性チェック、これは発生しないはず
			panic(fmt.Sprintf("charges > max? index :%v, frame : %v", c.Index, c.Core.F))
		}

		c.Core.Log.NewEventBuildMsg(glog.LogCooldownEvent, c.Index, a.String(), " cooldown ready").
			Write("type", a.String()).
			Write("charges_remain", c.AvailableCDCharge[a]).
			Write("cooldown_queue", c.cdQueueString(a))

		// キューにまだアイテムがあれば再度キューワーカーを開始
		if len(c.cdQueue) > 0 {
			c.startCooldownQueueWorker(a)
		}
	}

	c.cdCurrentQueueWorker[a] = &worker
	src = &worker

	// c.cooldownQueue[a][0] フレーム待ってからスタックを追加
	c.Core.Tasks.Add(worker, c.cdQueue[a][0])
}
