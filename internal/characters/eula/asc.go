package eula

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
)

// 氷潮の渦の長押しで冷酷な心スタックを2消費した場合、
// 砕けた光落の剣が即座に爆発し、
// 光界プロテクションで生成された光落の剣の基礎物理ダメージの50%を与える。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	// 長押しスキルのヒットラグ開始後かつ終了前に実行されるようにする
	// これにより長押しスキル終了後のヒットラグの影響を受けない
	aiA1 := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Icetide (Lightfall)",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		PoiseDMG:   180,
		Element:    attributes.Physical,
		Durability: 25,
		Mult:       burstExplodeBase[c.TalentLvlBurst()] * 0.5,
	}
	c.QueueCharTask(func() {
		c.Core.QueueAttack(
			aiA1,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), geometry.Point{Y: 2}, 6.5),
			a1Hitmark-(skillHoldHitmark+1),
			a1Hitmark-(skillHoldHitmark+1),
			c.burstStackCB,
		)
	}, skillHoldHitmark+1)
}

// 光界プロテクション発動時、氷潮の渦のCDがリセットされ、ユーラは冷酷な心を1スタック獲得する。
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}

	if c.grimheartStacks < 2 {
		c.grimheartStacks++
	}
	c.Core.Log.NewEvent("eula: grimheart stack", glog.LogCharacterEvent, c.Index).
		Write("current count", c.grimheartStacks)

	c.ResetActionCooldown(action.ActionSkill)
	c.Core.Log.NewEvent("eula a4 reset skill cd", glog.LogCharacterEvent, c.Index)
}
