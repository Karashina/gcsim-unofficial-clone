package diona

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
)

var burstFrames []int

const burstStart = 58 // 初撃

func init() {
	burstFrames = frames.InitAbilSlice(64) // Q -> N1/E
	burstFrames[action.ActionDash] = 43    // Q -> D
	burstFrames[action.ActionJump] = 44    // Q -> J
	burstFrames[action.ActionSwap] = 41    // Q -> Swap
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	// 初撃
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Signature Mix (Initial)",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Cryo,
		Durability: 25,
		Mult:       burst[c.TalentLvlBurst()],
	}
	c.Core.QueueAttack(ai, combat.NewCircleHit(c.Core.Combat.Player(), c.Core.Combat.PrimaryTarget(), nil, 3), 0, burstStart)

	// Tick処理
	ai.Abil = "Signature Mix (Tick)"
	ai.Mult = burstDot[c.TalentLvlBurst()]
	ap := combat.NewCircleHit(c.Core.Combat.Player(), c.Core.Combat.PrimaryTarget(), nil, 6.5)

	snap := c.Snapshot(&ai)
	hpplus := snap.Stats[attributes.Heal]
	maxhp := c.MaxHP()
	heal := burstHealPer[c.TalentLvlBurst()]*maxhp + burstHealFlat[c.TalentLvlBurst()]

	c.burstBuffArea = combat.NewCircleHitOnTarget(ap.Shape.Pos(), nil, 7)
	// 12.5秒持続する模様
	// TODO: フィールドが着地時に開始する前提（ゲーム内では動的）
	c.Core.Tasks.Add(func() {
		// 4凸チェック用の元素爆発ステータスを追加
		c.Core.Status.Add("diona-q", 750)
		// 2秒ごとにtick、最初のtickはt=2s（フィールド開始基準）、以降t=4,6,8,10,12; フィールド開始から12.5秒持続
		for i := 0; i < 6; i++ {
			c.Core.Tasks.Add(func() {
				// 攻撃
				c.Core.QueueAttackWithSnap(ai, snap, ap, 0)
				// 回復
				if !c.Core.Combat.Player().IsWithinArea(c.burstBuffArea) {
					return
				}
				c.Core.Player.Heal(info.HealInfo{
					Caller:  c.Index,
					Target:  c.Core.Player.Active(),
					Message: "Drunken Mist",
					Src:     heal,
					Bonus:   hpplus,
				})
			}, 120+i*120)
		}
		// 6凸
		if c.Base.Cons >= 6 {
			c.c6()
		}
	}, burstStart)

	// 1凸
	if c.Base.Cons >= 1 {
		// 終了後にエネルギー15回復、固定値でERの影響を受けない
		c.Core.Tasks.Add(func() {
			c.AddEnergy("diona-c1", 15)
		}, burstStart+750)
	}

	c.SetCDWithDelay(action.ActionBurst, 1200, 41)
	c.ConsumeEnergy(43)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstStart,
		State:           action.BurstState,
	}, nil
}
