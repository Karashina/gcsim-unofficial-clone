package core

import "github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"

func (c *Core) QueueAttackWithSnap(
	a combat.AttackInfo,
	s combat.Snapshot,
	p combat.AttackPattern,
	dmgDelay int,
	callbacks ...combat.AttackCBFunc,
) {
	if dmgDelay < 0 {
		panic("dmgDelay cannot be less than 0")
	}
	ae := combat.AttackEvent{
		Info:        a,
		Pattern:     p,
		Snapshot:    s,
		SourceFrame: c.F,
	}
	// nilでないコールバックのみ追加
	for _, f := range callbacks {
		if f != nil {
			ae.Callbacks = append(ae.Callbacks, f)
		}
	}
	c.queueDmg(&ae, dmgDelay)
}

func (c *Core) QueueAttackEvent(ae *combat.AttackEvent, dmgDelay int) {
	c.queueDmg(ae, dmgDelay)
}

func (c *Core) QueueAttack(
	a combat.AttackInfo,
	p combat.AttackPattern,
	snapshotDelay int,
	dmgDelay int,
	callbacks ...combat.AttackCBFunc,
) {
	// dmgDelay < snapshotDelay は起きてはならない。発生した場合は
	// キャラクターコードに問題がある
	if dmgDelay < snapshotDelay {
		panic("dmgDelay cannot be less than snapshotDelay")
	}
	if dmgDelay < 0 {
		panic("dmgDelay cannot be less than 0")
	}
	// AttackEvent を生成
	ae := combat.AttackEvent{
		Info:        a,
		Pattern:     p,
		SourceFrame: c.F,
	}
	// nilでないコールバックのみ追加
	for _, f := range callbacks {
		if f != nil {
			ae.Callbacks = append(ae.Callbacks, f)
		}
	}

	switch {
	case snapshotDelay < 0:
		// snapshotDelay < 0 はスナップショット不要を意味する。元素反応ダメージの最適化
		c.queueDmg(&ae, dmgDelay)
	case snapshotDelay == 0:
		c.generateSnapshot(&ae)
		c.queueDmg(&ae, dmgDelay)
	default:
		// タスクキューに追加。ここでの追跡は不要
		c.Tasks.Add(func() {
			c.generateSnapshot(&ae)
			c.queueDmg(&ae, dmgDelay-snapshotDelay)
		}, snapshotDelay)
	}
}

// このコードはcoreではなくplayerで処理すべきかもしれない
// queuedamage のラッパー的な便利関数であるため
//
// coreがチーム情報を持つのは適切か？おそらく不適切
func (c *Core) generateSnapshot(a *combat.AttackEvent) {
	a.Snapshot = c.Player.ByIndex(a.Info.ActorIndex).Snapshot(&a.Info)
}

func (c *Core) queueDmg(a *combat.AttackEvent, delay int) {
	if delay == 0 {
		c.Combat.ApplyAttack(a)
		return
	}
	c.Tasks.Add(func() {
		c.Combat.ApplyAttack(a)
	}, delay)
}
