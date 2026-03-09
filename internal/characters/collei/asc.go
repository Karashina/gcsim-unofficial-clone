package collei

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
)

const (
	sproutKey        = "collei-a1"
	sproutHitmark    = 86
	sproutTickPeriod = 89
	a4Key            = "collei-a4-modcheck"
)

// パーティメンバーがフローラルリングの帰還前に燃焼、激化、超激化、草激化、開花、超開花、烈開花の
// 元素反応を発動した場合、帰還時にキャラクターに「新芽」効果を付与し、3秒間
// コレイの攻撃力40%に相当する草元素ダメージを周囲の敵に継続的に与える。
// 初期効果の持続中に別の「新芽」効果が発動された場合、初期効果は除去される。
// 「新芽」によるダメージは元素スキルダメージとみなされる。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}

	//nolint:unparam // ignoring for now, event refactor should get rid of bool return of event sub
	f := func(...interface{}) bool {
		if c.sproutShouldProc {
			return false
		}
		if !c.StatusIsActive(skillKey) {
			return false
		}
		c.sproutShouldProc = true
		c.Core.Log.NewEvent("collei a1 proc", glog.LogCharacterEvent, c.Index)
		return false
	}

	for _, evt := range dendroEvents {
		switch evt {
		case event.OnHyperbloom, event.OnBurgeon:
			c.Core.Events.Subscribe(evt, f, "collei-a1")
		default:
			c.Core.Events.Subscribe(evt, func(args ...interface{}) bool {
				if _, ok := args[0].(*enemy.Enemy); !ok {
					return false
				}
				return f(args...)
			}, "collei-a1")
		}
	}
}

func (c *char) a1AttackInfo() combat.AttackInfo {
	return combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Floral Sidewinder (A1)",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagColleiSprout,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Dendro,
		Durability: 25,
		Mult:       0.4,
	}
}

// クイラーアンバーゾーン内のキャラクターが燃焼、激化、超激化、草激化、開花、超開花、烈開花の元素反応を発動した場合、
// ゾーンの持続時間が1秒延長される。
// 1回の切り札キティで最大3秒まで延長可能。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}

	//nolint:unparam // ignoring for now, event refactor should get rid of bool return of event sub
	f := func(args ...interface{}) bool {
		if !c.StatusIsActive(burstKey) {
			return false
		}
		atk := args[1].(*combat.AttackEvent)
		char := c.Core.Player.ByIndex(atk.Info.ActorIndex)
		if !char.StatusIsActive(a4Key) {
			return false
		}
		if c.burstExtendCount >= 3 {
			return false
		}
		c.ExtendStatus(burstKey, 60)
		c.burstExtendCount++
		c.Core.Log.NewEvent("collei a4 proc", glog.LogCharacterEvent, c.Index).
			Write("extend_count", c.burstExtendCount)
		return false
	}

	for _, evt := range dendroEvents {
		switch evt {
		case event.OnHyperbloom, event.OnBurgeon:
			c.Core.Events.Subscribe(evt, f, "collei-a4")
		default:
			c.Core.Events.Subscribe(evt, func(args ...interface{}) bool {
				if _, ok := args[0].(*enemy.Enemy); !ok {
					return false
				}
				return f(args...)
			}, "collei-a4")
		}
	}
}

func (c *char) a1Ticks(startFrame int, snap combat.Snapshot) {
	if !c.StatusIsActive(sproutKey) {
		return
	}
	if startFrame != c.sproutSrc {
		c.Core.Log.NewEvent("collei a1 tick ignored, src diff", glog.LogCharacterEvent, c.Index).
			Write("src", startFrame).
			Write("new src", c.sproutSrc)
		return
	}
	c.Core.QueueAttackWithSnap(
		c.a1AttackInfo(),
		snap,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 2),
		0,
	)
	c.Core.Tasks.Add(func() {
		c.a1Ticks(startFrame, snap)
	}, sproutTickPeriod)
}
