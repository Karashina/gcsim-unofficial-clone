package ayato

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var burstFrames []int

const (
	burstStart   = 101
	burstMarkKey = "ayato-burst-mark"
)

func init() {
	burstFrames = frames.InitAbilSlice(123) // Q -> N1
	burstFrames[action.ActionSkill] = 122   // Q -> E
	burstFrames[action.ActionDash] = 122    // Q -> D
	burstFrames[action.ActionJump] = 122    // Q -> J
	burstFrames[action.ActionSwap] = 120    // Q -> Swap
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Kamisato Art: Suiyuu",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
		Mult:       burst[c.TalentLvlBurst()],
	}

	// 円形成時にスナップショット（これは正しい？）
	var snap combat.Snapshot
	c.Core.Tasks.Add(func() { snap = c.Snapshot(&ai) }, burstStart)

	burstArea := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 10)
	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = burstatkp[c.TalentLvlBurst()]
	// burstStartから0.5秒ごとにティック
	for i := 0; i < 18*60; i += 30 {
		c.Core.Tasks.Add(func() {
			// 元素爆発のTick
			enemy := c.Core.Combat.RandomEnemyWithinArea(
				burstArea,
				func(e combat.Enemy) bool {
					return !e.StatusIsActive(burstMarkKey)
				},
			)
			var pos geometry.Point
			if enemy != nil {
				pos = enemy.Pos()
				enemy.AddStatus(burstMarkKey, 1.45*60, true) // 同じ敵は1.45秒間再ターゲットされない
			} else {
				pos = geometry.CalcRandomPointFromCenter(burstArea.Shape.Pos(), 1.5, 9.5, c.Core.Rand)
			}
			// 一定の遅延後にダメージを与える
			c.Core.QueueAttackWithSnap(ai, snap, combat.NewCircleHitOnTarget(pos, nil, 2.5), 38)

			// バフティック
			if !c.Core.Combat.Player().IsWithinArea(burstArea) {
				return
			}
			active := c.Core.Player.ActiveChar()
			active.AddAttackMod(character.AttackMod{
				Base: modifier.NewBaseWithHitlag("ayato-burst", 90),
				Amount: func(a *combat.AttackEvent, t combat.Target) ([]float64, bool) {
					return m, a.Info.AttackTag == attacks.AttackTagNormal
				},
			})
		}, i+burstStart)
	}

	if c.Base.Cons >= 4 {
		m := make([]float64, attributes.EndStatType)
		m[attributes.AtkSpd] = 0.15
		for _, char := range c.Core.Player.Chars() {
			char.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag("ayato-c4", 15*60),
				AffectedStat: attributes.AtkSpd,
				Amount: func() ([]float64, bool) {
					return m, true
				},
			})
		}
	}
	// シミュレーションにクールダウンを追加
	c.SetCD(action.ActionBurst, 20*60)
	// エネルギーを消費
	c.ConsumeEnergy(5)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}
