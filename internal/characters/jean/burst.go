package jean

import (
	"github.com/Karashina/gcsim-unofficial-clone/internal/frames"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/avatar"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attacks"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
)

var burstFrames []int

const burstStart = 40

func init() {
	burstFrames = frames.InitAbilSlice(90) // Q -> D/J
	burstFrames[action.ActionAttack] = 88  // Q -> N1
	burstFrames[action.ActionSkill] = 89   // Q -> E
	burstFrames[action.ActionSwap] = 88    // Q -> Swap
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	// pは敵がフィールドに出入りする回数
	enter := p["enter"]
	if enter < 1 {
		enter = 1
	}
	delay, ok := p["enter_delay"]
	if !ok {
		delay = 600 / enter
	}

	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Dandelion Breeze",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 50,
		Mult:       burst[c.TalentLvlBurst()],
	}
	snap := c.Snapshot(&ai)

	// 元素爆発開始後15フレームに初期ヒット
	c.Core.QueueAttackWithSnap(ai, snap, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 6), burstStart+15)

	// フィールドステータス
	c.Core.Status.Add("jean-q", 600+burstStart)

	// ユーザー指定のIn/Outダメージ量を処理
	// TODO: 移動に対応させる？
	ai.Abil = "Dandelion Breeze (In/Out)"
	ai.Mult = burstEnter[c.TalentLvlBurst()]
	// 最初の進入は元素爆発開始時
	for i := 0; i < enter; i++ {
		c.Core.QueueAttackWithSnap(ai, snap, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 6), burstStart+i*delay)
	}

	// フィールド消滅時のIn/Outダメージを処理
	c.Core.QueueAttackWithSnap(ai, snap, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 6), 600+burstStart)

	// 元素爆発開始時に回復
	hpplus := snap.Stats[attributes.Heal]
	atk := snap.Stats.TotalATK()
	heal := burstInitialHealFlat[c.TalentLvlBurst()] + atk*burstInitialHealPer[c.TalentLvlBurst()]
	healDot := burstDotHealFlat[c.TalentLvlBurst()] + atk*burstDotHealPer[c.TalentLvlBurst()]

	c.Core.Tasks.Add(func() {
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index,
			Target:  -1,
			Message: "Dandelion Breeze",
			Src:     heal,
			Bonus:   hpplus,
		})
	}, burstStart)

	self, ok := c.Core.Combat.Player().(*avatar.Player)
	if !ok {
		panic("target 0 should be Player but is not!!")
	}

	// 自身に攻撃
	selfSwirl := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Dandelion Breeze (Self Swirl)",
		Element:    attributes.Anemo,
		Durability: 25,
	}

	// 4凸は元素爆発開始直前にも１回適用
	if c.Base.Cons >= 4 {
		c.Core.Tasks.Add(func() {
			c.c4()
		}, burstStart-1)
	}
	c.burstArea = combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 6)
	// 持続時間は約10.6秒、最初のティックはフレーム100から開始、以降60フレームごと
	for i := 100; i <= 600+burstStart; i += 60 {
		c.Core.Tasks.Add(func() {
			if c.Core.Combat.Player().IsWithinArea(c.burstArea) {
				// 回復
				c.Core.Player.Heal(info.HealInfo{
					Caller:  c.Index,
					Target:  c.Core.Player.Active(),
					Message: "Dandelion Field",
					Src:     healDot,
					Bonus:   hpplus,
				})

				// 自己拡散
				ae := combat.AttackEvent{
					Info:        selfSwirl,
					Pattern:     combat.NewSingleTargetHit(0),
					SourceFrame: c.Core.F,
				}
				c.Core.Log.NewEvent("jean self swirling", glog.LogCharacterEvent, c.Index)
				self.ReactWithSelf(&ae)
			}
			// 4凸
			if c.Base.Cons >= 4 {
				c.c4()
			}
		}, i)
	}
	c.ConsumeEnergy(41)
	c.SetCDWithDelay(action.ActionBurst, 1200, 38)
	// 固有天賦2
	c.Core.Tasks.Add(func() {
		c.a4()
	}, burstStart+1)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // 最速キャンセル
		State:           action.BurstState,
	}, nil
}
