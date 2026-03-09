package character

import (
	"fmt"
	"math"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
)

func (c *CharWrapper) QueueCharTask(f func(), delay int) {
	if delay == 0 {
		f()
		return
	}
	c.queue.Add(f, delay)
}

func (c *CharWrapper) Tick() {
	// まず凍結時間を減少
	c.frozenFrames -= 1
	left := 0
	if c.frozenFrames < 0 {
		left = -c.frozenFrames
		c.frozenFrames = 0
	}
	// 残りがあれば経過時間を増加
	if left <= 0 {
		// このTickでは何もしない
		return
	}
	c.TimePassed += left

	// キャラキューに実行可能なアクションがあるか確認
	c.queue.Run()
}

func (c *CharWrapper) FramePausedOnHitlag() bool {
	return c.frozenFrames > 0
}

// ApplyHitlag はキャラクターに指定された期間のヒットラグを追加する
func (c *CharWrapper) ApplyHitlag(factor, dur float64) {
	// 凍結フレーム数 = 合計期間 * (1 - 係数)
	ext := int(math.Ceil(dur * (1 - factor)))
	c.frozenFrames += ext
	var logs []string
	var evt glog.Event
	if c.debug {
		logs = make([]string, 0, len(c.mods))
		evt = c.log.NewEvent(
			fmt.Sprintf("hitlag applied to char: %.3f", dur),
			glog.LogHitlagEvent, c.Index,
		).
			Write("duration", dur).
			Write("factor", factor).
			Write("frozen_frames", c.frozenFrames).
			SetEnded(*c.f + int(math.Ceil(dur)))
	}

	for i, v := range c.mods {
		if v.AffectedByHitlag() && v.Expiry() != -1 && v.Expiry() > *c.f {
			mod := c.mods[i]
			mod.Extend(mod.Key(), c.log, c.Index, ext)
			if c.debug {
				logs = append(logs, fmt.Sprintf("%v: %v", v.Key(), v.Expiry()))
			}
		}
	}

	if c.debug {
		evt.Write("mods affected", logs)
	}
}
