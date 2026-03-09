package chongyun

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

// 重華の氷螢之陽のフィールドが消えると、
// もう一本の霊刃が召喚され、付近の敵に重華の氷螢之陽のスキルダメージの100%の氷元素範囲ダメージを与える。
func (c *char) a4(delay, src int, useOldSnapshot bool) {
	if c.Base.Ascension < 4 {
		return
	}
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Spirit Blade: Chonghua's Layered Frost (A4)",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		PoiseDMG:   100,
		Element:    attributes.Cryo,
		Durability: 25,
		Mult:       skill[c.TalentLvlSkill()],
	}
	// スナップショットとスキル範囲の両方をタスククロージャにキャプチャする必要がある
	var snap combat.Snapshot
	if useOldSnapshot {
		snap = c.a4Snap
	} else {
		snap = c.Snapshot(&ai)
		c.a4Snap = snap
	}
	skillPattern := c.skillArea
	c.Core.Tasks.Add(func() {
		// srcが変更された場合、フィールドは既に変更済み
		if src != c.fieldSrc {
			return
		}
		enemy := c.Core.Combat.ClosestEnemyWithinArea(skillPattern, nil)
		var ap combat.AttackPattern
		if enemy != nil {
			ap = combat.NewCircleHitOnTarget(enemy, nil, 3.5)
		} else {
			ap = combat.NewCircleHitOnTarget(skillPattern.Shape.Pos(), nil, 3.5)
		}
		c.Core.QueueAttackWithSnap(ai, snap, ap, 0, c.a4CB, c.makeC4Callback())
	}, delay)
}

// この刃に命中した敵は氷元素耐性が10%低下する（8秒間）。
func (c *char) a4CB(a combat.AttackCB) {
	e, ok := a.Target.(*enemy.Enemy)
	if !ok {
		return
	}
	e.AddResistMod(combat.ResistMod{
		Base:  modifier.NewBaseWithHitlag("chongyun-a4", 480),
		Ele:   attributes.Cryo,
		Value: -0.10,
	})
}
