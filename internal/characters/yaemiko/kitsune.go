package yaemiko

import (
	"log"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
)

type kitsune struct {
	src         int
	deleted     bool
	kitsuneArea combat.AttackPattern
}

func (c *char) makeKitsune() {
	k := &kitsune{}
	k.src = c.Core.F
	k.deleted = false

	// プレイヤー位置に狐検知エリアを生成
	k.kitsuneArea = combat.NewCircleHitOnTarget(c.Core.Combat.Player().Pos(), nil, c.kitsuneDetectionRadius)

	// ティック開始
	c.Core.Tasks.Add(c.kitsuneTick(k), 120-skillStart)
	// タイムアウト時（他で削除されていない場合）にこれを削除するタスクを追加
	c.Core.Tasks.Add(func() {
		// ここでは.deletedをチェックするだけで良いと思う
		if k.deleted {
			return
		}
		// これで削除可能
		c.popOldestKitsune()
	}, 900-skillStart) // e ani + duration

	if len(c.kitsunes) == 0 {
		c.Core.Status.Add(yaeTotemStatus, 900-skillStart)
	}
	// 最も古いものを先にポップ
	if len(c.kitsunes) == 3 {
		c.popOldestKitsune()
	}
	c.kitsunes = append(c.kitsunes, k)
	c.SetTag(yaeTotemCount, c.sakuraLevelCheck())
}

func (c *char) popAllKitsune() {
	for i := range c.kitsunes {
		c.kitsunes[i].deleted = true
	}
	c.kitsunes = c.kitsunes[:0]
	c.Core.Status.Delete(yaeTotemStatus)
	c.SetTag(yaeTotemCount, 0)
}

func (c *char) popOldestKitsune() {
	if len(c.kitsunes) == 0 {
		// ポップするものがない？
		return
	}

	c.kitsunes[0].deleted = true
	c.kitsunes = c.kitsunes[1:]

	// ここでステータスを確認
	if len(c.kitsunes) > 0 {
		dur := c.Core.F - c.kitsunes[0].src + (900 - skillStart)
		if dur < 0 {
			log.Panicf("oldest totem should have expired already? dur: %v totem: %v", dur, *c.kitsunes[0])
		}
		c.Core.Status.Add(yaeTotemStatus, dur)
	} else {
		c.Core.Status.Delete(yaeTotemStatus)
	}

	c.SetTag(yaeTotemCount, len(c.kitsunes))
}

func (c *char) kitsuneBurst(ai combat.AttackInfo, pattern combat.AttackPattern) {
	for i := 0; i < c.sakuraLevelCheck(); i++ {
		c.Core.QueueAttack(ai, pattern, burstThunderbolt1Hitmark+i*24, burstThunderbolt1Hitmark+i*24)
		if c.Base.Cons >= 1 {
			c.Core.Tasks.Add(func() {
				c.AddEnergy("yae-c1", 8)
			}, burstThunderbolt1Hitmark+i*24)
		}
		c.a1()
		c.Core.Log.NewEvent("sky kitsune thunderbolt", glog.LogCharacterEvent, c.Index).
			Write("src", c.kitsunes[i].src).
			Write("delay", burstThunderbolt1Hitmark+i*24)
	}
	c.popAllKitsune()
}

func (c *char) kitsuneTick(totem *kitsune) func() {
	return func() {
		// 削除済みの場合は何もしない
		if totem.deleted {
			return
		}
		// 6凸
		// 殺生櫻はレベル2で開始。最大レベルは4に増加し、攻撃は敵の防御力の45%を無視する。

		lvl := c.sakuraLevelCheck() - 1
		if c.Base.Cons >= 2 {
			lvl += 1
		}

		ai := combat.AttackInfo{
			Abil:       "Sesshou Sakura Tick",
			ActorIndex: c.Index,
			AttackTag:  attacks.AttackTagElementalArt,
			Mult:       skill[lvl][c.TalentLvlSkill()],
			ICDTag:     attacks.ICDTagElementalArt,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Electro,
			Durability: 25,
		}

		c.Core.Log.NewEvent("sky kitsune tick at level", glog.LogCharacterEvent, c.Index).
			Write("sakura level", lvl+1)

		var c4cb combat.AttackCBFunc
		if c.Base.Cons >= 4 {
			done := false
			c4cb = func(a combat.AttackCB) {
				if a.Target.Type() != targets.TargettableEnemy {
					return
				}
				if done {
					return
				}
				done = true
				c.c4()
			}
		}
		if c.Base.Cons >= 6 {
			ai.IgnoreDefPercent = 0.60
		}

		// 攻撃を1回生成
		// 優先度: 敵 > ガジェット
		tick := func(pos geometry.Point) {
			c.Core.QueueAttack(
				ai,
				combat.NewCircleHitOnTarget(pos, nil, 0.5),
				1,
				1,
				c.particleCB,
				c4cb,
			)
		}

		// まず敵をターゲットする
		enemy := c.Core.Combat.RandomEnemyWithinArea(totem.kitsuneArea, nil)
		if enemy != nil {
			tick(enemy.Pos())
		} else {
			// 敵がターゲットされなかった場合はガジェットをターゲット
			gadget := c.Core.Combat.RandomGadgetWithinArea(totem.kitsuneArea, nil)
			if gadget != nil {
				tick(gadget.Pos())
			}
		}

		// 約2.9秒ごとにティック
		c.Core.Tasks.Add(c.kitsuneTick(totem), 176)
	}
}

func (c *char) sakuraLevelCheck() int {
	count := len(c.kitsunes)
	if count < 0 {
		// これはトーテムがない場合のベースケース（そうでなければ6凸の場合1になる）
		return 0
	}
	if count > 3 {
		panic("wtf more than 3 totems")
	}
	return count
}
