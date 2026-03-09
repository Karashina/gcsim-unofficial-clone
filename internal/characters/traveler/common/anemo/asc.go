package anemo

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
)

// 通常攻撃コンボの最終段で風刃を放ち、経路上の全敵に攻撃力の60%の風元素ダメージを与える。
func (c *Traveler) a1() {
	if c.Base.Ascension < 1 || c.NormalCounter != c.NormalHitNum-1 {
		return
	}
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Slitting Wind (A1)",
		AttackTag:  attacks.AttackTagNormal,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupPoleExtraAttack,
		StrikeType: attacks.StrikeTypeSlash,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       0.6,
	}
	c.QueueCharTask(func() {
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHit(
				c.Core.Combat.Player(),
				c.Core.Combat.PrimaryTarget(),
				nil,
				1,
			),
			0,
			0,
		)
	}, a1Hitmark[c.gender])
}

const a4ICDKey = "traveleranemo-a4-icd"

// 風の巻で敵を倒すと5秒間HPが2%回復する。
// この効果は5秒ごとに1回のみ発動。
func (c *Traveler) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	c.Core.Events.Subscribe(event.OnTargetDied, func(args ...interface{}) bool {
		if _, ok := args[0].(*enemy.Enemy); !ok {
			return false
		}
		atk := args[1].(*combat.AttackEvent)
		if atk.Info.ActorIndex != c.Index {
			return false
		}
		if atk.Info.AttackTag != attacks.AttackTagElementalArt && atk.Info.AttackTag != attacks.AttackTagElementalArtHold {
			return false
		}
		if c.StatusIsActive(a4ICDKey) {
			return false
		}

		c.AddStatus(a4ICDKey, 300, true)

		for i := 0; i < 5; i++ {
			c.QueueCharTask(func() {
				c.Core.Player.Heal(info.HealInfo{
					Caller:  c.Index,
					Target:  c.Index,
					Message: "Second Wind",
					Type:    info.HealTypePercent,
					Src:     0.02,
				})
			}, (i+1)*60) // 回復は死亡1秒後に開始
		}

		return false
	}, "traveleranemo-a4")
}
