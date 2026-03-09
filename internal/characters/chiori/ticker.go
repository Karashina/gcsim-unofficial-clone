package chiori

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
)

func (c *char) kill(t *ticker) {
	if t != nil {
		t.kill()
	}
}

// 人形用の汎用ティッカー
type ticker struct {
	c     *core.Core
	alive bool

	cb       func()
	interval int

	onDeath func()
	queuer
}

type queuer func(cb func(), delay int)

// kill は既存のティッカーの動作を停止する
func (g *ticker) kill() {
	g.alive = false
	g.cb = nil
	g.interval = 0
	if g.onDeath != nil {
		g.onDeath()
	}
}

func newTicker(c *core.Core, life int, q queuer) *ticker {
	// life <= 0 かどうかはチェックしない
	// life が <= 0 の場合、次のタスクチェック時に
	// ガジェットが自己破壊する
	g := &ticker{
		alive:  true,
		c:      c,
		queuer: q,
	}
	if g.queuer == nil {
		g.queuer = c.Tasks.Add
	}
	g.queuer(func() {
		if !g.alive {
			return
		}
		g.kill()
	}, life)
	return g
}

func (g *ticker) tick() {
	// ガジェットが死んでいる場合は何もしない
	if !g.alive {
		return
	}
	// コールバックを実行
	if g.cb != nil {
		g.cb()
	}
	// 次のアクションをキューに入れる
	if g.interval > 0 {
		g.queuer(g.tick, g.interval)
	}
}
